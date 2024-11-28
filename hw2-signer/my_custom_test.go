package main

import (
	"fmt"
	"testing"
)

func TestSingleHash(t *testing.T) {
	inn := [2]int{0, 1}
	out := make(chan interface{}, 1)
	in := make(chan interface{}, 1)

	go SingleHash(in, out)

	out <- inn[0]
	out <- inn[1]

	close(out)
	for e := range in {
		fmt.Println(e)
	}
}
func TestExecutePipelineSingleHash(t *testing.T) {
	inn := []int{0, 1}

	myFlowJobs := []job{
		job(func(in, out chan interface{}) {
			for e := range inn {
				out <- e
			}
			close(out)
		}),
		job(SingleHash),
		job(func(in, out chan interface{}) {
			for e := range in {
				fmt.Println(e)
			}
		}),
	}

	ExecutePipeline(myFlowJobs...)
}
