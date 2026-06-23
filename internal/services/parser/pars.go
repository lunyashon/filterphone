package parser

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/lib/auth"
	"github.com/lunyashon/filterphone/internal/lib/curl"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"github.com/lunyashon/filterphone/internal/services/phsearch"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (csp *CSVParser) parseCsv(c *gin.Context) {

	valid, err := auth.ValidateToken(c, csp.cfg)
	if err != nil {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": err.Error()})
		return
	}
	if !valid {
		c.JSON(structure.Status[codes.Unauthenticated], gin.H{"error": "unauthorized"})
		return
	}
	var (
		errg, ctxg = errgroup.WithContext(c.Request.Context())
		phones     map[string]string
		exclude    map[string]string
		blacklist  map[string]string
	)

	fh, err := c.FormFile("filter")
	if err != nil {
		c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
		return
	}

	errg.Go(func() error {
		phones, err = csp.parseFile(ctxg, fh)
		if err != nil {
			return err
		}
		return nil
	})

	if c.PostForm("use_exclude") == "1" {
		excludeF, err := c.FormFile("exclude")
		if err != nil {
			c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
			return
		}
		errg.Go(func() error {
			exclude, err = csp.parseFile(ctxg, excludeF)
			if err != nil {
				return err
			}
			return nil
		})
	}

	if c.PostForm("use_blacklist") == "1" {
		blacklistF, err := c.FormFile("blacklist")
		if err != nil {
			c.JSON(structure.Status[codes.InvalidArgument], gin.H{"error": err.Error()})
			return
		}
		errg.Go(func() error {
			blacklist, err = csp.parseFile(ctxg, blacklistF)
			if err != nil {
				return err
			}
			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	filteredPhones, err := csp.filterPhones(phones, exclude, blacklist)
	if err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	var (
		numbers       = make(map[string]structure.Numbers)
		errg1, ctxg1  = errgroup.WithContext(c.Request.Context())
		mu            sync.Mutex
		dupl          = c.PostForm("use_duplicate")
		numbersFailed = make(map[string]structure.Numbers)
	)

	errg1.SetLimit(20)

	for phone := range filteredPhones {
		errg1.Go(func() error {
			abc, tail, err := phsearch.ParsingPhone(phone)
			if err != nil {
				return err
			}
			number, err := csp.numbers.GetNumbers(ctxg1, int16(abc), tail)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil
				}
				return err
			}
			if dupl == "1" {
				r, err := csp.checkDuplicateFromZarya(ctxg1, phone)
				if err != nil {
					csp.log.Error("failed to check duplicate from zarya", "error", err)
					mu.Lock()
					numbersFailed[phone] = *number
					mu.Unlock()
					return nil
				}
				if r {
					mu.Lock()
					numbersFailed[phone] = *number
					mu.Unlock()
					return nil
				}
			}
			mu.Lock()
			numbers[phone] = *number
			mu.Unlock()
			return nil
		})
	}

	if err := errg1.Wait(); err != nil {
		c.JSON(structure.Status[codes.Internal], gin.H{"error": err.Error()})
		return
	}

	c.JSON(structure.Status[codes.OK], gin.H{"phones": numbers, "failed": numbersFailed})
}

func (csp *CSVParser) parseFile(ctx context.Context, fh *multipart.FileHeader) (map[string]string, error) {

	if fh.Size > 5<<20 {
		return nil, status.Errorf(codes.InvalidArgument, "file is too large (max 5MB)")
	}

	f, err := fh.Open()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to open file")
	}
	defer f.Close()

	r := csv.NewReader(f)

	r.Comma = ';'
	r.LazyQuotes = true

	var (
		phones = make(map[string]string)
	)

	for {
		select {
		case <-ctx.Done():
			return nil, status.Errorf(codes.Canceled, "context canceled")
		default:
		}
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, status.Errorf(codes.Internal, "failed to read file: %v", err)
		}
		if len(rec) == 0 {
			break
		}

		if rec[0] == "" {
			continue
		}

		isPhone := true
		for _, ch := range rec[0] {
			if !unicode.IsDigit(ch) {
				isPhone = false
				break
			}
		}
		if !isPhone {
			continue
		}

		if _, ok := phones[rec[0]]; ok {
			continue
		}

		phones[rec[0]] = rec[0]
	}

	return phones, nil
}

func (csp *CSVParser) filterPhones(
	phones map[string]string,
	exclude map[string]string,
	blacklist map[string]string,
) (map[string]string, error) {

	for phone := range phones {
		if _, ok := exclude[phone]; ok {
			delete(phones, phone)
			continue
		}
		if _, ok := blacklist[phone]; ok {
			delete(phones, phone)
			continue
		}
	}
	return phones, nil
}

type ZaryaCheckDuplicate struct {
	Phone string `json:"phone"`
}

type ZaryaCheckDuplicateResponse struct {
	Result bool `json:"result"`
}

func (csp *CSVParser) checkDuplicateFromZarya(ctx context.Context, phone string) (bool, error) {
	client := http.Client{Timeout: 60 * time.Second}
	request := ZaryaCheckDuplicate{Phone: phone}
	param, err := json.Marshal(request)
	if err != nil {
		return false, err
	}
	const maxRetries = 3

	var statusCode int
	for i := 0; i <= maxRetries; i++ {
		body, statusCode, err := curl.CurlWithContext(ctx, &client, csp.cfg.B24ZaryaUrl, "POST", bytes.NewReader(param), nil)
		if err != nil {
			if i < maxRetries && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
				time.Sleep(time.Duration((i+1)*(i+1)) * time.Second)
				continue
			}
			return false, err
		}

		if statusCode >= 500 && i < maxRetries {
			time.Sleep(time.Duration((i+1)*(i+1)) * time.Second)
			continue
		}

		var response ZaryaCheckDuplicateResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return false, err
		}

		if statusCode != 200 {
			return false, fmt.Errorf("failed to check duplicate from zarya: %d", statusCode)
		}

		return response.Result, nil
	}

	return false, fmt.Errorf("failed to check duplicate from zarya after retries: %d", statusCode)
}
