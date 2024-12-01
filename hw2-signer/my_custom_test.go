package main

import (
	"fmt"
	"testing"
)

func TestSingleHash(t *testing.T) {
	const N = 2
	inputs := [N]int{0, 1}
	out := make(chan interface{}, N)
	in := make(chan interface{}, N)

	go SingleHash(in, out)

	for _, v := range inputs {
		out <- v
	}
	close(out)

	for result := range in {
		fmt.Println(result)
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

func TestExecutePipelineMultiHash(t *testing.T) {
	inn := []int{0, 1}

	myFlowJobs := []job{
		job(func(in, out chan interface{}) {
			for e := range inn {
				out <- e
			}
			close(out)
		}),
		job(MultiHash),
		job(func(in, out chan interface{}) {
			for e := range in {
				fmt.Println(e)
			}
		}),
	}

	ExecutePipeline(myFlowJobs...)
}
