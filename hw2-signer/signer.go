package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

func swapChans(in, out chan interface{}) (chan interface{}, chan interface{}) {
	return out, in
}

// ExecutePipeline которая обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
// Написание нескольких функций, которые считают нам какую-то условную хеш-сумму от входных данных
func ExecutePipeline(workers ...job) {
	fmt.Println("ExecutePipeline go")
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)

	wg := sync.WaitGroup{}

	for _, w := range workers {
		wg.Wait()
		wg.Add(1)
		w(in, out)
		wg.Done()
		//in, out = swapChans(in, out)
	}
}

// SingleHash считает значение crc32(data)+"~"+crc32(md5(data)) (конкатенация двух строк через ~),
// где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {
	//fmt.Println("SingleHash go")
	wg := sync.WaitGroup{}
	for e := range out {
		dataInt := e.(int)
		data := strconv.Itoa(dataInt)

		chan1crc32 := make(chan string, 1)
		go func(data string, out chan string) {
			out <- DataSignerCrc32(data)
			close(out)
		}(data, chan1crc32)

		md5ch := make(chan string, 1)
		wg.Wait()
		wg.Add(1)
		go func(data string, out chan string) {
			out <- DataSignerMd5(data)
			close(out)
			wg.Done()
		}(data, md5ch)

		chan2crc32 := make(chan string, 1)
		go func(data string, out chan string) {
			out <- DataSignerCrc32(data)
			close(out)
		}(<-md5ch, chan2crc32)

		//fmt.Println(fir, sec)
		in <- <-chan1crc32 + "~" + <-chan2crc32
	}
	close(in)
	//fmt.Println("SingleHash end")
}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
// где th=0..5 (т.е. 6 хешей на каждое входящее значение),
// потом берёт конкатенацию результатов в порядке расчета (0..5),
// где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {
	fmt.Println("MultiHash go")

	var temp string
	th := []string{"0", "1", "2", "3", "4"}

	for data := range out {
		temp = ""
		for _, t := range th {
			fir := DataSignerCrc32(t + data.(string))
			fmt.Println(t, fir)
			temp += fir
		}
		in <- temp
	}
	close(in)
	fmt.Println("MultiHash end")
}

// CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/),
// объединяет отсортированный результат через _ (символ подчеркивания) в одну строку
func CombineResults(in, out chan interface{}) {
	fmt.Println("Combic go")
	arr := make([]string, 0, 20)
	for v := range in {
		arr = append(arr, v.(string))
	}
	sort.Strings(arr)
	for i := 0; i < len(arr)-1; i++ {
		out <- arr[i] + "_"
	}
	out <- arr[len(arr)-1]
	defer close(out)
}
