package validatorx

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestIdCard(t *testing.T) {

	type Test struct {
		IdCard string `validate:"id_card"`
	}

	type Test15 struct {
		IdCard string `validate:"id_card_15"`
	}

	t.Run(`valid_18_len_id_card`, func(t *testing.T) {
		test := Test{}
		test.IdCard = `510802199409214133`
		err := New().Struct(test)
		if err != nil {
			t.Error(err)
		}
	},
	)

	t.Run(`valid_15len`, func(t *testing.T) {
		test := Test15{}
		test.IdCard = `510802950901131`
		err := New().Struct(test)
		if err != nil {
			t.Error(err)
		}
	},
	)

	t.Run(`invalid_18len_brith_day_part`, func(t *testing.T) {
		test1 := Test{}
		test1.IdCard = `510802177009014131`
		err := New().Struct(test1)
		if err == nil {
			t.Error(`510802177009014131 should be invalid`)
		}

		test2 := Test{}
		test2.IdCard = `510802199513014131`
		err = New().Struct(test2)
		if err == nil {
			t.Error(`510802199513014131 should be invalid`)
		}

		test3 := Test{}
		test3.IdCard = `510802199502314139`
		err = New().Struct(test3)
		if err == nil {
			t.Error(`510802199502314139 should be invalid`)
		}
	},
	)

	t.Run(`invalid_18len_last_index `, func(t *testing.T) {
		test := Test{}
		test.IdCard = `510802199409214138`
		err := New().Struct(test)
		if err == nil {
			t.Error(`51080219940921413x should be invalid`)
		}
	},
	)
}

// 生成身份证18位
func gen18LenIdCard() string {
	province := []string{
		"11", "12", "13", "14", "15", "21", "22", "23", "31", "32", "33", "34", "35",
		"36", "37", "41", "42", "43", "44", "45", "46", "50", "51", "52", "53", "54",
		"61", "62", "63", "64", "65", "71", "81", "82",
	}

	randomIdCard := province[rand.Intn(len(province)-1)] + randLengthNum(4) + randBirthDay() + randLengthNum(3)
	result := getIdCard(randomIdCard)

	return result
}

func randLengthNum(length int) (randNum string) {
	for i := 0; i < length; i++ {
		randNum += strconv.Itoa(rand.Intn(9))
	}

	return
}

func randBirthDay() string {
	randomTime := rand.Int63n(time.Now().Unix()-5608000) + 5608000
	randomNow := time.Unix(randomTime, 0)
	month := strconv.Itoa(int(randomNow.Month()))
	year := strconv.Itoa(randomNow.Year())
	day := strconv.Itoa(randomNow.Day())

	if len(month) == 1 {
		month = "0" + month
	}

	if len(day) == 1 {
		day = "0" + day
	}

	return year + month + day
}

// getLastNum 获取18位身份证号码的最后一位合法
func getIdCard(idCard string) string {
	// 根据 2^(index-1) % 11 得出位置:权重 哈希表 index从右至左算
	weightList := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2, 1}
	// 校验码值换算表
	checkMap := map[int]int{0: 1, 2: 10, 3: 9, 4: 8, 5: 7, 6: 6, 7: 5, 8: 4, 9: 3, 10: 2, 1: 0}

	var count int
	for index, value := range idCard {
		count += int(value-'0') * weightList[index]
	}

	for k, v := range checkMap {
		checkNum := checkMap[(count+k)%11]
		if checkNum == v {
			if k == 10 {
				idCard = idCard + "X"
			} else {
				idCard = idCard + strconv.Itoa(k)
			}

			break
		}
	}

	return idCard
}
