# containers-from-scratch

## Write a easy Docker engine from scratch! This issue includes:

- Part 1: Modifying the hostname of a container using Sethostname
- Part 2: Using CLONE_NEWUTS to isolate host and container UTS Namespace
- Part 3: Using CLONE_NEWPID to isolate processes & modify the container root directory
- Part 4: Isolating Mount Namespace with CLONE_NEWNS
- Part 5: Use cgroup to limit the resources available to the container



​	You need root permissions for this version to work. Or you can adapt it to be a rootless container by as shown in [these slides](https://speakerdeck.com/lizrice/rootless-containers-from-scratch). 

​	Note that the Go code uses some syscall definitions that are only available when building with GOOS=linux.

​	rootfs package is available at https://cdimage.ubuntu.com/ubuntu-base/releases/14.04/release/ubuntu-base-14.04.1-core-arm64.tar.gz