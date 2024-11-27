package main

import (
	"fmt"
	"sort"
)

func swapChans(in, out chan interface{}) (chan interface{}, chan interface{}) {
	return out, in
}

// ExecutePipeline которая обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
// Написание нескольких функций, которые считают нам какую-то условную хеш-сумму от входных данных
func ExecutePipeline(workers ...job) {
	fmt.Println("ExecutePipeline go")
	in := make(chan interface{})
	out := make(chan interface{})

	for _, w := range workers {
		w(in, out)
		in, out = swapChans(in, out)
	}
}

// SingleHash считает значение crc32(data)+"~"+crc32(md5(data)) (конкатенация двух строк через ~),
// где data - то что пришло на вход (по сути - числа из первой функции)
func SingleHash(in, out chan interface{}) {
	fmt.Println("SingleHash go")
	for data := range in {
		fir := DataSignerCrc32(data.(string))
		sec := DataSignerCrc32(DataSignerMd5(data.(string)))

		fmt.Println(fir, sec)
		out <- fir + "~" + sec
	}
}

// MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой к строке и строки),
// где th=0..5 (т.е. 6 хешей на каждое входящее значение),
// потом берёт конкатенацию результатов в порядке расчета (0..5),
// где data - то что пришло на вход (и ушло на выход из SingleHash)
func MultiHash(in, out chan interface{}) {
	fmt.Println("MultiHash go")
	var temp string
	th := []string{"0", "1", "2", "3", "4"}
	for data := range in {
		temp = ""
		for _, t := range th {
			fir := DataSignerCrc32(t + data.(string))
			fmt.Println(t, fir)
			temp += fir
		}
		out <- temp
	}
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
}
