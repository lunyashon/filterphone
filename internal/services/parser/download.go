package parser

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/lunyashon/filterphone/internal/lib/curl"
	"golang.org/x/sync/errgroup"
)

var (
	urlOpendata = "https://zarya.apcheit.ru/api/v1/opendata.download?file=%s"
)

func (csp *CSVParser) DownloadOpendata() error {

	var (
		files = []string{
			"ABC-3xx",
			"ABC-4xx",
			"ABC-8xx",
			"DEF-9xx",
		}

		errg, ctxg = errgroup.WithContext(context.Background())
	)

	if err := os.MkdirAll("files", 0o755); err != nil {
		return err
	}

	client := http.Client{Timeout: 60 * time.Second}

	for _, file := range files {
		file := file // capture range variable for goroutine
		errg.Go(func() error {
			body, statusCode, err := curl.CurlWithContext(
				ctxg,
				&client,
				fmt.Sprintf(urlOpendata, file),
				"GET",
				nil,
				nil,
			)
			if err != nil {
				return err
			}
			if statusCode != 200 {
				return fmt.Errorf("failed to download file: %d", statusCode)
			}
			targetPath := filepath.Join("./files", file+".csv")
			if _, err := os.Stat(targetPath); err == nil {
				if err := os.Remove(targetPath); err != nil {
					fmt.Println("failed to remove file: ", err)
				}
			}
			if err := os.WriteFile(targetPath, body, 0o644); err != nil {
				return err
			}

			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		return err
	}

	return nil
}
