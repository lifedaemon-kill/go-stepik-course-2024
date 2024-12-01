package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

// ExecutePipeline которая обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
// Написание нескольких функций, которые считают нам какую-то условную хеш-сумму от входных данных
func ExecutePipeline(workers ...job) {
	fmt.Println("ExecutePipeline go")
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)

	//wg := sync.WaitGroup{}

	for _, w := range workers {
		//wg.Wait()
		//wg.Add(1)
		w(in, out)
		//wg.Done()

		out = in
		in = make(chan interface{}, 100)
	}
}

// SingleHash считает значение crc32(data)+"~"+crc32(md5(data)) (конкатенация двух строк через ~),
// где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {
	for e := range out {

		dataInt := e.(int)
		data := strconv.Itoa(dataInt)

		chCrc1 := make(chan string)
		chCrc2 := make(chan string)

		go func() {
			chCrc1 <- DataSignerCrc32(data)
		}()

		go func() {
			md5Data := DataSignerMd5(data)
			chCrc2 <- DataSignerCrc32(md5Data)
		}()

		cr1 := <-chCrc1
		cr2 := <-chCrc2

		in <- cr1 + "~" + cr2
	}
	close(in)
}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
// где th=0..5 (т.е. 6 хешей на каждое входящее значение),
// потом берёт конкатенацию результатов в порядке расчета (0..5),
// где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {
	fmt.Println("MultiHash go")

	th := []string{"0", "1", "2", "3", "4"}
	wg := sync.WaitGroup{}
	in = make(chan interface{}, 100)

	for dataRaw := range out {
		data := strconv.Itoa(dataRaw.(int))

		arr := [5]string{}
		for i, t := range th {
			wg.Add(1)
			go func(index int, t string) {
				arr[index] = DataSignerCrc32(t + data)
				wg.Done()
			}(i, t)

		}
		wg.Wait()

		in <- arr[0] + arr[1] + arr[2] + arr[3] + arr[4]
	}
	close(in)
	fmt.Println("MultiHash end")
}

// CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/),
// объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {
	fmt.Println("Combic go")
	arr := make([]string, 0, 100)
	in = make(chan interface{}, 100)

	for v := range out {
		arr = append(arr, v.(string))
	}

	sort.Strings(arr)
	line := ""
	for i := 0; i < len(arr)-1; i++ {
		line += arr[i] + "_"
	}
	line += arr[len(arr)-1]
	in <- line
	close(in)
}
