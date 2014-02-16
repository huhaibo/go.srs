package main

import (
	"fmt"
	"reflect"
)

type BlackWinlin struct {
	id int
}
type RedWinlin struct {
	name string
}

func main() {
	bw := BlackWinlin{id:10}

	var rtmp_pkt *RedWinlin = nil
	fmt.Println("rtmp==========================")

	rtmp_pkt = nil
	if my_rtmp_expect(&bw, &rtmp_pkt) {
		fmt.Println("discoveryed pkt from black:", rtmp_pkt)
	}

	fmt.Println()
	rtmp_pkt = nil
	if my_rtmp_expect(&RedWinlin{}, &rtmp_pkt) {
		fmt.Println("discoveryed pkt from red:", rtmp_pkt)
	}

	fmt.Println()
	fmt.Println("rtmp==========================")
	var src_black_pkt *BlackWinlin = &bw
	var src_red_pkt *RedWinlin = &RedWinlin{name: "hello"}

	rtmp_pkt = nil
	if my_rtmp_expect(&src_black_pkt, &rtmp_pkt) {
		fmt.Println("discoveryed pkt from black:", rtmp_pkt)
	}

	fmt.Println()
	rtmp_pkt = nil
	if my_rtmp_expect(&src_red_pkt, &rtmp_pkt) {
		fmt.Println("discoveryed pkt from red:", rtmp_pkt)
	}

	fmt.Println()
	fmt.Println("rtmp==========================")
	// set the value which is ptr to ptr
	var prtmp_pkt **RedWinlin = nil
	if my_rtmp_expect(&src_red_pkt, prtmp_pkt) {
		fmt.Println("discoveryed pkt from red(ptr):", prtmp_pkt)
		fmt.Println("discoveryed pkt from red(value):", *prtmp_pkt)
	}
	prtmp_pkt = &rtmp_pkt
	if my_rtmp_expect(&src_red_pkt, prtmp_pkt) {
		fmt.Println("discoveryed pkt from red(ptr):", prtmp_pkt)
		fmt.Println("discoveryed pkt from red(value):", *prtmp_pkt)
	}
}

func my_rtmp_expect(src interface {}, expect interface {}) (ok bool){
	ok = false

	src_rt := reflect.TypeOf(src)
	src_rv := reflect.ValueOf(src)
	src_ptr_rt := reflect.PtrTo(src_rt)
	expect_rt := reflect.TypeOf(expect)
	expect_rv := reflect.ValueOf(expect)

	if expect_rv.Kind() != reflect.Ptr || expect_rv.IsNil() {
		fmt.Println("expect must be ptr and not nil")
		return
	}

	fmt.Println("type info, src:", src_rt, "ptr(src):", src_ptr_rt, ", expect:", expect_rt)
	fmt.Println("value info, src:", src_rv, ", src.Elem():", src_rv.Elem(), ", expect:", expect_rv, ", expect.Elem():", expect_rv.Elem())
	fmt.Println("convertible src=>expect:", src_rt.ConvertibleTo(expect_rt))
	fmt.Println("ptr convertible ptr(src)=>expect:", src_ptr_rt.ConvertibleTo(expect_rt))
	fmt.Println("elem convertible src=>expect.Elem()", src_rt.ConvertibleTo(expect_rt.Elem()))
	fmt.Println("settable src:", src_rv.CanSet(), ", expect:", expect_rv.CanSet())
	fmt.Println("elem settable src:", src_rv.Elem().CanSet(), ", expect:", expect_rv.Elem().CanSet())

	if expect_rv.Elem().CanSet() {
		if src_rt.ConvertibleTo(expect_rt){
			fmt.Println("directly match, src=>expect")
			expect_rv.Elem().Set(src_rv.Elem())
			ok = true
			return
		}

		expect_elem_rt := expect_rt.Elem()
		if src_rt.ConvertibleTo(expect_elem_rt) {
			fmt.Println("pointer match, src=>*expect")
			expect_rv.Elem().Set(src_rv)
			ok = true
			return
		}
		fmt.Println("not match, donot set expect")
	} else {
		fmt.Println("expect cannot set")
	}

	return
}
