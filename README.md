# 第六届“强网杯”全国网络安全挑战赛 线上赛 PWN pokemongo

## 版权

~~赛题版权归出题人（bibi?）和关联方所有，侵删。~~ EDIT: 见[#原题](#原题)

题解及相关代码如需转载，请先获得我的许可。

## 赛题

附件：`pokemongo.zip`[百度盘](https://pan.baidu.com/s/13HPccqRbvr9CjOxdCPV3xQ)，提取码`GAME`

## 题解

`pokemongo`代码量不大，但是IDA尚不支持`ABIInternal`寄存器传参，得用`__usercall`手工修正后才能看。（或许有好用的工具只是我不了解）

大白话：给我一段Go代码，我帮你编译、执行，唯一的要求是代码中不能`import`，请开始你的表演，获取flag文件内容。

详细流程如下：
* `main()`
  * 输入长度
  * 分多行输入字符串内容（其实`bufio.Scanner`是多余的，base64解码时会忽略换行符）
  * 对输入的字符串解base64编码，得到源码
  * `sanitizeAndRun(src_code)`
    * `sanitize(src_code)`
      * 调用`go/parser`将源码解析为AST
      * 检查AST不包含`import`声明
      * 调用`go/printer`将AST还原为源码
      * 返回处理后的源码
    * `run(sanitized_src_code)`
      * 创建临时目录
      * 将源码保存到临时目录下的`main.go`
      * 开始计时5秒
      * 执行`go build -buildmode=pie <源码路径>`编译
      * 执行编译产物，捕获标准输出与标准错误的内容
      * 返回编译产物的输出
    * 返回`run()`的结果
  * 输出`sanitizeAndRun()`的结果

先浏览一遍《The Go Programming Language Specification》[built-in函数列表](https://go.dev/ref/spec#Built-in_functions)，感觉只有`print`/`println`函数能用的上。

不允许`import`，也就没法使用`//go:linkname`，只能自己造漏洞来利用。

Go的主要贡献者之一[Russ Cox曾经转发过](https://twitter.com/_rsc/status/1403845276033486850)这样一段话：
> Go loses its memory safety guarantees if you write concurrent software. Rust loses its memory safety guarantees if you use non-trivial data structures. C++ loses its memory safety guarantees if you use pointers (or references).

思路可以确定下来，靠race实现类型混淆，达成任意内存读写。

`interface`的存储结构定义如下：
```go
type iface struct {
	tab  *itab
	data unsafe.Pointer
}

type eface struct {
	_type *_type
	data  unsafe.Pointer
}
```
Go将没有方法的`interface`实现为`eface`，有方法的`interface`实现为`iface`。
对于我们来说`eface`更直接，更方便利用。

我们可以用一个goroutine对`interface{}`来回赋两种不同类型的值，另一个goroutine去观测它。
`eface`内两个指针的赋值，既不是原子读写，也没有加锁，因此可以观察到中间状态，实现类型混淆。
通过generics语法，我们可以方便地混淆任意类型而不用复制粘贴代码。（之所以把`OutputType`放在`InputType`前面，是为了利用类型推断，方便使用）
```go
func typeConfuse[OutputType, InputType any](input *InputType) (output *OutputType) {
	var intf any
	stop := false
	go func() {
		for !stop {
			intf = any(input)
			intf = any(output)
		}
	}()
	for {
		if ptr, ok := intf.(*OutputType); ok && ptr != nil {
			stop = true
			return ptr
		}
	}
}
```

使用样例：
```go
// copied from reflect.SliceHeader
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func main() {
	slice := []byte("Hello World")
	_slice := typeConfuse[SliceHeader](&slice)
	_slice.Len = 5
	println(string(slice))
}
```
[在Go Playground上执行这段代码](https://go.dev/play/p/ma2dzWAnPBY)可以看到输出为`Hello`，我们可以自由修改`SliceHeader`进行任意内存读写。

以上步骤相当于拿到了`unsafe.Pointer`，之后的利用步骤有多种思路：
1. 利用`runtime.open`/`runtime.read`读取flag文件
2. 利用`runtime.mmap`写入shellcode
3. 利用`runtime/internal/syscall.Syscall6`执行任意syscall（execve）

比赛时不确定flag文件路径，没有选择第1条思路。
`pokemongo`已经给我们提供了回显，因此思路2相比思路3没有优越性。
最终选择了思路3。

我们可以很方便定位到`runtime/internal/syscall.Syscall6`，但是传参的时候尴尬了。
`pokemongo`使用Go 1.18.0编译（可以用`go version -m pokemongo`查看），大量runtime函数还在用Go祖传的`ABI0`栈传参，而我们在Go代码中声明的函数“指针”全都是`ABIInternal`寄存器传参。
就算我们只用`syscall;ret`这样的gadget，`rdx`寄存器一般用于传递闭包的上下文，我们在Go代码中不太方便控制`execve`的`envp`参数。

好在workaround也很单纯：声明函数类型的时候，前9个参数占满寄存器传参，后面的参数自然就用栈传递了，可以很方便地模拟`ABI0`的参数。
```go
func (_0,_1,_2,_3,_4,_5,_6,_7,_8 uintptr, syscall_nr uintptr, filename *byte, argv **byte, envp **byte)
```

完整利用代码见本仓库`exploit`目录。

## 源码

本仓库`src/pokemongo.go`为逆向得到的源码，运行`src/build.sh`再次构建可以得到（除`.note.go.buildid`以外）与赛题完全一致的可执行文件。
看似短短100行代码，想做到完全一致还是有一定难度的，可以加深对Go的理解。

## 原题

后来注意到本题并非原创，原题是[Google CTF 2019 Finals的Gomium Browser](https://github.com/google/google-ctf/tree/master/2019/finals/pwn-gomium)。变更有：
* 大幅度简化了交互
* 检查`import`的时候不再允许`fmt`包
* `go build`参数追加了`-buildmode=pie`

[原题作者的博客](https://blog.stalkr.net/2019/12/the-gomium-browser-exploits.html)也很值得一读，学习其它思路。
