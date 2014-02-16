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

func my_rtmp_expect(pkt interface {}, v interface {}) (ok bool){
	/*
    func my_rtmp_expect(pkt interface {}, v interface {}){
        rt := reflect.TypeOf(v)
        rv := reflect.ValueOf(v)
        
        // check the convertible and convert to the value or ptr value.
        // for example, the v like the c++ code: Msg**v
        pkt_rt := reflect.TypeOf(pkt)
        if pkt_rt.ConvertibleTo(rt){
            // directly match, the pkt is like c++: Msg**pkt
            // set the v by: *v = *pkt
            rv.Elem().Set(reflect.ValueOf(pkt).Elem())
            return
        }

        if pkt_rt.ConvertibleTo(rt.Elem()) {
            // ptr match, the pkt is like c++: Msg*pkt
            // set the v by: *v = pkt
            rv.Elem().Set(reflect.ValueOf(pkt))
            return
        }
    }
	 */
	ok = false

	pkt_rt := reflect.TypeOf(pkt)
	pkt_rv := reflect.ValueOf(pkt)
	pkt_ptr_rt := reflect.PtrTo(pkt_rt)
	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		fmt.Println("expect must be ptr and not nil")
		return
	}

	fmt.Println("type info, src:", pkt_rt, "ptr(src):", pkt_ptr_rt, ", expect:", rt)
	fmt.Println("value info, src:", pkt_rv, ", src.Elem():", pkt_rv.Elem(), ", expect:", rv, ", expect.Elem():", rv.Elem())
	fmt.Println("convertible src=>expect:", pkt_rt.ConvertibleTo(rt))
	fmt.Println("ptr convertible ptr(src)=>expect:", pkt_ptr_rt.ConvertibleTo(rt))
	fmt.Println("elem convertible src=>expect.Elem()", pkt_rt.ConvertibleTo(rt.Elem()))
	fmt.Println("settable src:", pkt_rv.CanSet(), ", expect:", rv.CanSet())
	fmt.Println("elem settable src:", pkt_rv.Elem().CanSet(), ", expect:", rv.Elem().CanSet())

	// check the convertible and convert to the value or ptr value.
	// for example, the v like the c++ code: Msg**v
	if rv.Elem().CanSet() {
		if pkt_rt.ConvertibleTo(rt){
			// directly match, the pkt is like c++: Msg**pkt
			// set the v by: *v = *pkt
			fmt.Println("directly match, src=>expect")
			rv.Elem().Set(pkt_rv.Elem())
			ok = true
			return
		}

		if pkt_rt.ConvertibleTo(rt.Elem()) {
			// ptr match, the pkt is like c++: Msg*pkt
			// set the v by: *v = pkt
			fmt.Println("pointer match, src=>*expect")
			rv.Elem().Set(pkt_rv)
			ok = true
			return
		}
		fmt.Println("not match, donot set expect")
	} else {
		fmt.Println("expect cannot set")
	}

	return
}
