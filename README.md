go.srs
======

[SRS(simple-rtmp-server)](https://github.com/winlinvip/simple-rtmp-server) written by google go language.

### 产品定位

GO.SRS和SRS的定位不太一样，可以互补。<br/>
SRS主要是RTMP源站，外加HLS和转码，单进程（通过转码消耗系统能力）。<br/>
GO.SRS主要是流服务器，包括源站和边缘，支持RTMP/RTMPT/RTMPE/RTMPS/HLS/DASH/RTMFP/RTSP等，支持多进程。<br/>
如下图所示：
<pre>
+---------------------------+    +------------------------------+
|     GO.SRS(流服务器)      +-->-+  SRS转码/Chnvideo编转码集群  |
|   (IPv4/IPv6/TCP/UDP)     |    +------------------------------+
| (源站/边缘/单进程/多进程) |                                    
| (RTMP/RTMPE/RTMPT/RTMPS)  |    +------------------------------+
|   (HTTP/HLS/HDS/DASH)     |-->-+  Chnvideo收录/时移/播放器    |
|     (RTSP,RTMFP)          |    +------------------------------+
+---------------------------+                                    
IPv4/IPv6: 同时支持IPv4/IPv6。
TCP：支持基于TCP的协议，譬如RTMP和HTTP系列。
UDP：支持基于UDP的协议，譬如RTMFP等。
源站和边缘：支持集群，譬如RTMP系列的源站和边缘，HTTP只需要支持源站（边缘用NGINX等成熟方案）。
RTMP系列：流服务器的基础协议。
RTSP系列：支持RTSP流协议，支持一些摄像头。
RTMFP：Adobe的FlashP2P方案。
SRS转码：SRS可以用ffmpeg转码，为转码的开源方案。
Chnvideo：商业方案
</pre>

### GO开发环境

参考[http://golang.org/doc/install](http://golang.org/doc/install)<br/>
<strong>Step 1:</strong>下载GO<br/>
https://code.google.com/p/go/downloads/list<br/>
<strong>Step 2:</strong>解压GO<br/>
<pre>
tar xf go1.2.linux-amd64.tar.gz
sudo ln -sf `pwd`/go /usr/local/go
</pre>
<strong>Step 3:</strong>设置环境变量<br/>
<pre>
export GOROOT=/usr/local/go
export PATH=$PATH:$GOROOT/bin
</pre>
注意：所有环境变量的设置可以编辑/etc/profile

### 编译方法(Build)

<strong>Step 1:</strong> set GOPATH if not set<br/>
<pre>
export GOPATH=~/mygo
</pre>
<strong>Step 2:</strong> get and build srs<br/>
<pre>
go get github.com/winlinvip/go.srs/go_srs
</pre>
<strong>Step 3:</strong> start SRS <br/>
<pre>
$GOPATH/bin/go_srs
</pre>
注意：编译出来的go_srs不依赖于GO开发环境，可以独立部署。

Beijing, 2014<br/>
Winlin

