package parser

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lunyashon/filterphone/internal/lib/structure"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
)

func (csp *CSVParser) Restore(c *gin.Context) {

	go func() {

		var (
			errg, ctxg = errgroup.WithContext(context.Background())
		)

		if err := csp.numbers.DeleteNumbers(ctxg); err != nil {
			csp.log.Error("failed to delete numbers", "error", err)
			return
		}

		fileName := []string{
			"./files/ABC-3xx.csv",
			"./files/ABC-4xx.csv",
			"./files/ABC-8xx.csv",
			"./files/DEF-9xx.csv",
		}

		for _, fileName := range fileName {
			n := fileName
			errg.Go(func() error {
				if _, err := os.Stat(n); os.IsNotExist(err) {
					csp.log.Error("file not found", "error", err)
					return err
				}

				return csp.parseCSVFile(ctxg, n)
			})
		}

		if err := errg.Wait(); err != nil {
			csp.log.Error("failed to restore numbers", "error", err)
			return
		}
	}()

	c.JSON(structure.Status[codes.OK], gin.H{"message": "started"})
}

func (csp *CSVParser) parseCSVFile(ctx context.Context, file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	csvReader.Comma = ';'
	csvReader.LazyQuotes = true

	for {
		rec, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if len(rec) == 0 {
			break
		}
		if rec[1] == "От" {
			continue
		}

		fmt.Printf("%+v\n", rec)

		if rec[0] == "" {
			rec[0] = "0"
		}
		code, err := strconv.Atoi(rec[0])
		if err != nil {
			return err
		}

		if rec[1] == "" {
			rec[1] = "0"
		}
		from, err := strconv.Atoi(rec[1])
		if err != nil {
			return err
		}

		if rec[2] == "" {
			rec[2] = "0"
		}
		to, err := strconv.Atoi(rec[2])
		if err != nil {
			return err
		}

		if rec[3] == "" {
			rec[3] = "0"
		}
		capacity, err := strconv.Atoi(rec[3])
		if err != nil {
			return err
		}

		if rec[7] == "" {
			rec[7] = "0"
		}
		inn, err := strconv.ParseInt(rec[7], 10, 64)
		if err != nil {
			return err
		}
		numbers := structure.Numbers{
			Code:      int16(code),
			From:      from,
			To:        to,
			Capacity:  capacity,
			Operator:  rec[4],
			Region:    rec[5],
			Territory: rec[6],
			INN:       inn,
		}
		err = csp.numbers.CreateNumbers(ctx, numbers)
		if err != nil {
			return err
		}
	}

	return nil
}
