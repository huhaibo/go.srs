package main

import (
	"bytes"
	"fmt"
	amf "github.com/hydra13142/encoding/AMF"
	"github.com/hydra13142/flv"
	"os"
)

func MergeFlv(files []string, file string) {

	flvs := make([]*flv.Flv, len(files))
	for i := 0; i < len(flvs); i++ {
		flvs[i] = flv.New()
	}

	for i, name := range files {
		r, err := os.Open(name)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = flvs[i].ReadFrom(r)
		if err != nil {
			fmt.Println(err)
			return
		}
		r.Close()
	}

	r := amf.New()
	script := flvs[0].Tags[0].Data
	_, script, _ = r.DecodeAMF0(script)
	info, script, _ := r.DecodeAMF0(script)
	meta := info.(amf.ECMA)
	var step int
	l := len(flvs[0].Tags)
	/*
	   i,n:=0,0
	   a,b:=0,0
	   for ;i<l;i++{
	       if flvs[0].Tags[i].Video() {
	           break
	       }
	   }
	   a=i
	   for l--;i<l;i++{
	       if flvs[0].Tags[i].Video() {
	           b=i
	           n++
	       }
	   }
	   step = (flvs[0].Tags[b].Video()-flvs[0].Tags[a].Video())/n
	*/
	if rate, ok := meta.Get("framerate").(float64); ok {
		step = int(1000 / rate)
	} else {
		var a [2]int
		for i, j := 1, 0; j < 2 && i < l; i++ {
			if flvs[0].Tags[i].Video() {
				a[j] = i
				j++
			}
		}
		step = flvs[0].Tags[a[1]].Time() - flvs[0].Tags[a[0]].Time()
	}
	place := []int{}
	times := []int{}
	for i := 1; i < len(flvs); i++ {
		l += len(flvs[i].Tags) - 1
	}
	tag := make([]flv.Tag, l)
	copy(tag, flvs[0].Tags)
	l = len(flvs[0].Tags)
	d, v, a, p := 0, 0, 0, 0
	for i := 1; i < l; i++ {
		if tag[i].Video() {
			if flv.Keyframe(tag[i].Data[0]) {
				place = append(place, p)
				times = append(times, tag[i].Time())
			}
			v += tag[i].Size()
		} else if tag[i].Audio() {
			a += tag[i].Size()
		} else if tag[i].Script() {
			d += tag[i].Size()
		}
		p += tag[i].Size() + 15
	}
	for _, one := range flvs[1:] {
		x, y := l-1, 0
		for !tag[x].Video() {
			x--
		}
		for !one.Tags[y].Video() {
			y++
		}
		move := tag[x].Time() - one.Tags[y].Time() + step
		for i := 1; i < len(one.Tags); i, l = i+1, l+1 {
			tag[l] = one.Tags[i]
			tag[l].SetTime(tag[l].Time() + move)
			if tag[l].Video() {
				if flv.Keyframe(tag[l].Data[0]) {
					place = append(place, p)
					times = append(times, tag[l].Time())
				}
				v += tag[i].Size()
			} else if tag[i].Audio() {
				a += tag[i].Size()
			} else if tag[i].Script() {
				d += tag[i].Size()
			}
			p += tag[l].Size() + 15
		}
	}
	num := len(times)
	if tag[l-1].Time()-tag[l-2].Time() > step {
		lasttime := tag[l-2].Time() + step
		tag[l-1].SetTime(lasttime)
		if tag[l-1].Video() {
			if flv.Keyframe(tag[l-1].Data[0]) {
				times[num-1] = lasttime
			}
		}
	}
	r.Reset()
	meta.Del("keyframes")
	remain, _ := r.CountAMF0(meta)
	header := len(flvs[0].Head) + 4
	offset := 11 + 13 + (remain + 18*num + 47) + 3 + 4
	positions := make([]interface{}, num)
	timestamps := make([]interface{}, num)
	for i := 0; i < num; i++ {
		positions[i] = float64(place[i] + offset)
		timestamps[i] = float64(times[i])
	}
	duration := float64(tag[l-1].Time()) / 1000
	meta.Set("duration", duration)
	meta.Set("lasttimestamp", duration)
	frame := amf.NewECMA()
	frame.Set("filepositions", amf.Strict(positions))
	frame.Set("times", amf.Strict(timestamps))
	meta.Set("keyframes", amf.Object{"", frame})
	meta.Set("filesize", float64(p+offset+header))
	meta.Set("datasize", float64(d))
	meta.Set("audiosize", float64(a))
	meta.Set("videosize", float64(v))
	meta.Set("lastkeyframelocation", positions[num-1])
	meta.Set("lastkeyframetimestamp", timestamps[num-1])
	sx, _ := r.EncodeAMF0("onMetaData")
	sy, _ := r.EncodeAMF0(meta)
	tag[0].Data = bytes.Join([][]byte{sx, sy, []byte{0, 0, 9}}, nil)
	tag[0].SetSize(len(tag[0].Data))
	flvs[0].Tags = tag

	w, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	flvs[0].WriteTo(w)
	w.Close()
}

func main() {
	if l := len(os.Args); l <= 1 {
		fmt.Println("usage : executable [src...] dest")
		return
	} else if l == 2 {
		MergeFlv(os.Args[1:], os.Args[1])
	} else {
		l--
		MergeFlv(os.Args[1:l], os.Args[l])
	}
}
