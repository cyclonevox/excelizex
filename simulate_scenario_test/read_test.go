package simulateTest

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cyclonevox/excelizex"
)

// 该单元测试仅作为模拟业务文件模拟
func Test_Batch_Read(t *testing.T) {
	f, err := os.Open("./xlsx/batch_data.xlsx")
	if err != nil {
		panic(err)
	}
	excel := excelizex.New(f)
	r := excel.Read(new(batchData), "Sheet1").SetValidates(newValidation()).
		SetConvert("id-string", func(rawData string) (any, error) {
			for strings.HasPrefix(rawData, "0") {
				rawData = strings.TrimPrefix(rawData, "0")
			}

			return rawData, nil
		})

	t.Run("each_business_cost_1s_1_goroutine", func(t *testing.T) {
		_, err = r.Run(func(any any) (err error) {
			b := any.(*batchData)
			time.Sleep(1 * time.Second)
			fmt.Println(b)

			return
		}, 1)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("each_business_cost_1s_10_goroutine", func(t *testing.T) {
		_, err = r.Run(func(any any) (err error) {
			b := any.(*batchData)
			time.Sleep(1 * time.Second)
			fmt.Println(b)

			return
		}, 10)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("each_business_cost_200ms_10_goroutine", func(t *testing.T) {
		var rr *excelizex.Result
		rr, err = r.Run(func(any any) (err error) {
			b := any.(*batchData)
			time.Sleep(200 * time.Millisecond)
			fmt.Println(b)

			return
		}, 10)

		if err != nil {
			t.Error(err)
		}

		fmt.Println(rr.TotalNum())
		fmt.Println(rr.ErrorNum())
	})

	t.Run("each_business_cost_200ms_100_goroutine", func(t *testing.T) {
		var rr *excelizex.Result
		if rr, err = r.Run(func(any any) (err error) {
			b := any.(*batchData)
			time.Sleep(200 * time.Millisecond)
			fmt.Println(b)

			return errors.New("test")
		}, 100); err != nil {
			t.Error(err)
		}

		fmt.Println(rr.TotalNum())
		fmt.Println(rr.ErrorNum())
	})
}
