package main

import (
	"fmt"
	"reflect"
)

type Winlin interface {
	Demo(int) (string)
}
type BlackWinlin struct {
	id int
}
func (o *BlackWinlin) Demo(n int) (string) {
	return fmt.Sprintf("black#%v, n=%v", o.id, n)
}
type RedWinlin struct {
}
func (o *RedWinlin) Demo(n int) (string) {
	return fmt.Sprintf("red, n=%v", n)
}

func show(v interface {}) {
	var o0 Winlin = v.(Winlin)
	fmt.Println(o0.Demo(10))
}
func show1(v interface {}) {
	var o0 Winlin = v.(*BlackWinlin)
	fmt.Println(o0.Demo(10))
}
func show2(v interface {}) {
	var o0 Winlin = v.(*RedWinlin)
	fmt.Println(o0.Demo(10))
}
func reflect0(i interface {}) {
	fmt.Println("reflect demo:", i)

	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	fmt.Printf("t=%v, v=%v\n", t, v)
}
func expect(v interface {}, t interface {}) {
	rv := reflect.ValueOf(t)
	rt := reflect.TypeOf(t)
	fmt.Printf("expect value to interface, t=%v, rt=%v, rv=%v\n", t, rt, rv)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		fmt.Println("expect must be ptr and not nil")
		return
	}

	fmt.Println("can set t:", rv.CanSet())
	rve := rv.Elem()
	fmt.Println("can set te:", rve.CanSet())
	rve.Set(reflect.ValueOf(v))
}
func expect1(from Winlin, to *Winlin) {
	tov := reflect.ValueOf(to)
	tov.Elem().Set(reflect.ValueOf(from))
}
func expect2(from Winlin, to *Winlin) {
	*to = from
}
func my_pi(v interface {}) {
	fmt.Println("winlin pi=", v.(string))
}
func main() {
	my_pi("str")

	bw := BlackWinlin{id:10}
	show(&bw)
	show1(&bw)
	// panic: interface conversion: interface is *main.BlackWinlin, not *main.RedWinlin
	//show2(&bw)

	reflect0(bw)
	reflect0(&bw)

	var w Winlin = &RedWinlin{}
	expect(&bw, &w)
	fmt.Println("expect ret=", w.Demo(20))

	w = &RedWinlin{}
	expect1(&bw, &w)
	fmt.Println("expect1 ret=", w.Demo(30))

	w = &RedWinlin{}
	expect2(&bw, &w)
	fmt.Println("expect2 ret=", w.Demo(40))

	mb := &BlackWinlin{id:21}
	fmt.Println("test new black, ", mb.Demo(10))
	expect(&bw, &mb)
	fmt.Println("expect ret=", mb.Demo(40))

	mb = &BlackWinlin{id:21}
	w = mb
	fmt.Println("test new black, ", mb.Demo(10))
	expect1(&bw, &w)
	fmt.Println("expect1 ret=", mb.Demo(40))

	var rtmp_pkt *RedWinlin
	fmt.Println("rtmp==========================")
	my_rtmp_expect(&bw, &rtmp_pkt)
	fmt.Println("discoveryed pkt from black:", rtmp_pkt)
	fmt.Println()
	my_rtmp_expect(&RedWinlin{}, &rtmp_pkt)
	fmt.Println("discoveryed pkt from red:", rtmp_pkt)

	fmt.Println("rtmp==========================")
	var src_black_pkt *BlackWinlin = &bw
	var src_red_pkt *RedWinlin = &RedWinlin{}
	my_rtmp_expect(&src_black_pkt, &rtmp_pkt)
	fmt.Println("discoveryed pkt from black:", rtmp_pkt)
	fmt.Println()
	my_rtmp_expect(&src_red_pkt, &rtmp_pkt)
	fmt.Println("discoveryed pkt from red:", rtmp_pkt)
}

func my_rtmp_expect(src interface {}, expect interface {}) {
	src_rt := reflect.TypeOf(src)
	src_rv := reflect.ValueOf(src)
	src_ptr_rt := reflect.PtrTo(src_rt)
	expect_rt := reflect.TypeOf(expect)
	expect_rv := reflect.ValueOf(expect)
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
			return
		}
		expect_elem_rt := expect_rt.Elem()
		if src_rt.ConvertibleTo(expect_elem_rt) {
			fmt.Println("pointer match, src=>*expect")
			expect_rv.Elem().Set(src_rv)
			return
		}
		fmt.Println("not match, donot set expect")
	} else {
		fmt.Println("expect cannot set")
	}
}
