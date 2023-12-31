package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

//docker run <cmd> <arguments>
//docker child <cmd> <arguments>

type FileSystem struct {
	Source string
	Target string
	Fstype string
	Flags  uintptr
}

func main() {

	if len(os.Args) <= 1 {
		panic("no enough arguments")
	}

	if len(os.Args) <= 2 {
		panic("please specify the command to run")
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("help")
	}
}

func run() {

	fmt.Printf("parent pid: %d\n", os.Getpid())
	//fmt.Printf("Running %v \n", os.Args[2:])

	args := append([]string{"child"}, os.Args[2:]...)

	/*proc目录是所有进程的元数据存放地方
	我们的二进制文件也会出现在这里
	下面这行代码会在新创建的容器内执行child函数，
	proc/self/exe是一个特殊的文件，包含当前可执行文件的内存映像。
	换句话说，会让进程重新运行自己，但是传递child作为第一个参数。
	这个可执行程序让我们能够执行另一个程序，执行一个由用户请求的程序（由‘os.Args[2:]’中定义的内容）。
	基于这个简单的结构，我们就能够创建一个容器。*/
	cmd := exec.Command("/proc/self/exe", args...)

	// 将操作系统标准io重定向到容器中
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置一些系统进程属性，下面这行代码负责创建一个新的独立进程
	// 创建进程或容器来运行我们提供的命令
	// CLONE_NEWUTS运行容器有独立的UTS
	// CLONE_NEWPID为新的命名空间进程提供pids
	// CLONE_NEWNS为mount提供新的命名空间
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		// systemd中的挂载会递归共享属性。
		//取消对新挂载命名空间的递归共享属性。
		//它阻止与主机共享新的命名空间。
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

func child() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, "doing unmount")
		}
		must(syscall.Unmount("proc", syscall.MNT_DETACH))
		must(syscall.Unmount("mytemp", syscall.MNT_DETACH))
	}()

	fmt.Printf("child pid: %d\n", os.Getpid())
	//fmt.Printf("Running %v \n", os.Args[2:])

	//Set cgroup values
	cg()

	must(syscall.Sethostname([]byte("container")))

	//must(syscall.Mount("ubuntu-fs", "ubuntu-fs", "", syscall.MS_BIND, ""))
	//must(os.MkdirAll("ubuntu-fs/old-ubuntu-fs", 0700))
	//must(syscall.PivotRoot("ubuntu-fs", "ubuntu-fs/old-ubuntu-fs"))
	//must(os.Chdir("/"))

	//set the child process's root file.   use pivot_root
	//You have to download the os rootfs
	must(syscall.Chroot("/root/docker-tutorial/ubuntu-fs"))
	must(os.Chdir("/"))

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Current working directory:", currentDir)
	}

	//when we execute ps command,it will read /proc directory. However, the ubuntu rootfs
	//dont have /proc, so we have to mount the /proc in host machine.

	fs := FileSystem{
		Source: "proc",
		Target: "proc",
		Fstype: "proc",
		Flags:  uintptr(syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV),
	}
	must(syscall.Mount(fs.Source, fs.Target, fs.Fstype, fs.Flags, ""))

	fs.Source = "thing"
	fs.Target = "mytemp"
	fs.Fstype = "tmpfs"
	//mytemp是相对路径，绝对路径其实是/root/docker-tutorial/ubuntu-fs/mytemp，所以需要确保该目录下有该目录
	must(syscall.Mount(fs.Source, fs.Target, fs.Fstype, fs.Flags, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())

	//must(syscall.Unmount("ubuntu-fs", 0))

}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	//  /sys/fs/cgroup/pids/container
	err := os.Mkdir(filepath.Join(pids, "container"), 0755)
	if err != nil {
		fmt.Println("container exist,its ok")
	}
	must(os.WriteFile(filepath.Join(pids, "container/pids.max"), []byte("10"), 0700))
	// Removes the new cgroup in place after the container exits
	must(os.WriteFile(filepath.Join(pids, "container/notify_on_release"), []byte("1"), 0700))
	must(os.WriteFile(filepath.Join(pids, "container/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
