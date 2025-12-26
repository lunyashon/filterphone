package parser

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"io"
	"mime/multipart"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/lib/auth"
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
		numbers      = make(map[string]structure.Numbers)
		errg1, ctxg1 = errgroup.WithContext(c.Request.Context())
		mu           sync.Mutex
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

	c.JSON(structure.Status[codes.OK], gin.H{"phones": numbers})
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
