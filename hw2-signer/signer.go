package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ExecutePipeline которая обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
// Написание нескольких функций, которые считают нам какую-то условную хеш-сумму от входных данных
func ExecutePipeline(workers ...job) {
	//fmt.Println("ExecutePipeline go")
	in := make(chan interface{})
	var out chan interface{}
	wg := &sync.WaitGroup{}

	for _, w := range workers {
		out = make(chan interface{})

		wg.Add(1)
		go func(currentJob job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)

			currentJob(in, out)
		}(w, in, out)

		in = out
	}

	wg.Wait()
	//fmt.Println("ExecutePipeline end")
}

// SingleHash считает значение crc32(data)+"~"+crc32(md5(data)) (конкатенация двух строк через ~),
// где data - то что пришло на вход (по сути - числа из первой функции)

func SingleHash(in, out chan interface{}) {
	start := time.Now()

	md5Mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for e := range in {
		wg.Add(1)

		go func(e interface{}) {
			defer wg.Done()

			data := strconv.Itoa(e.(int))

			chCrc1 := make(chan string)
			chCrc2 := make(chan string)
			md5 := make(chan string)

			go func() {
				chCrc1 <- DataSignerCrc32(data)
			}()

			go func() {
				md5Mutex.Lock()
				md5Chan := DataSignerMd5(data)
				md5Mutex.Unlock()
				md5 <- md5Chan
			}()
			go func() {
				md5Data := <-md5
				chCrc2 <- DataSignerCrc32(md5Data)
			}()

			crc32 := <-chCrc1
			md5Crc32 := <-chCrc2

			out <- crc32 + "~" + md5Crc32
		}(e)

	}

	wg.Wait()
	end := time.Since(start)
	fmt.Println("single end =", end)
}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
// где th=0..5 (т.е. 6 хешей на каждое входящее значение),
// потом берёт конкатенацию результатов в порядке расчета (0..5),
// где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {
	start := time.Now()

	th := []string{"0", "1", "2", "3", "4", "5"}
	wg := sync.WaitGroup{}
	wgGlobal := &sync.WaitGroup{}

	for dataRaw := range in {
		wgGlobal.Add(1)

		go func(dataRaw interface{}, wg *sync.WaitGroup) {
			defer wgGlobal.Done()

			data := dataRaw.(string)

			arr := make([]string, len(th))

			for i, t := range th {
				wg.Add(1)

				go func(index int, t string) {
					defer wg.Done()
					arr[index] = DataSignerCrc32(t + data)
				}(i, t)

			}
			wg.Wait()
			out <- strings.Join(arr, "")
		}(dataRaw, &wg)

	}
	wgGlobal.Wait()

	end := time.Since(start)
	fmt.Println("multi end =", end)
}

// CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/),
// объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {

	arr := make([]string, 0, 100)

	for v := range in {
		arr = append(arr, v.(string))
	}

	sort.Strings(arr)
	line := ""
	for i := 0; i < len(arr)-1; i++ {
		line += arr[i] + "_"
	}
	line += arr[len(arr)-1]

	out <- line
}
