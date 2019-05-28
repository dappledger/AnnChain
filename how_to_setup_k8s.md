安装 kubelet kubeadm kubectl
```
cat <<EOF > /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
setenforce 0
yum install -y kubelet kubeadm kubectl
systemctl enable kubelet && systemctl start kubelet
```

安装docker

```
curl -fsSL get.docker.com -o get-docker.sh
sudo sh get-docker.sh —mirror Aliyun
# 安装完成后将 用户 添加到docker 组，执行以下命令，更新缓存！
sudo usermod -aG docker $USER
```

启动docker

```
systemctl daemon-reload
systemctl enable docker
systemctl start docker
```

**由于安装k8s使用的官方镜像地址k8s.grc.io被封，所以需要使用国内的docker镜像**

首先查看要使用的镜像： 
```
kubeadm config images list

k8s.gcr.io/kube-apiserver:v1.13.4
k8s.gcr.io/kube-controller-manager:v1.13.4
k8s.gcr.io/kube-scheduler:v1.13.4
k8s.gcr.io/kube-proxy:v1.13.4
k8s.gcr.io/pause:3.1
k8s.gcr.io/etcd:3.2.24
k8s.gcr.io/coredns:1.2.6

```

然后使用脚本下载：

```
images=(
    kube-apiserver:v1.13.4
    kube-controller-manager:v1.13.4
    kube-scheduler:v1.13.4
    kube-proxy:v1.13.4
    pause:3.1
    etcd:3.2.24
    coredns:1.2.6
)

for imageName in ${images[@]} ; do
    docker pull registry.cn-hangzhou.aliyuncs.com/google_containers/$imageName
    docker tag registry.cn-hangzhou.aliyuncs.com/google_containers/$imageName k8s.gcr.io/$imageName
    docker rmi registry.cn-hangzhou.aliyuncs.com/google_containers/$imageName
done

＃启动master
swapoff -a
echo 1 > /proc/sys/net/bridge/bridge-nf-call-iptables
export KUBE_REPO_PREFIX="registry-vpc.cn-beijing.aliyuncs.com/bbt_k8s"
export KUBE_ETCD_IMAGE="registry-vpc.cn-beijing.aliyuncs.com/bbt_k8s/etcd-amd64:3.0.17"
kubeadm init —kubernetes-version=v1.13.4 —pod-network-cidr=192.168.0.0/16
sudo cp /etc/kubernetes/admin.conf $HOME/
sudo chown $(id -u):$(id -g) $HOME/admin.conf
export KUBECONFIG=$HOME/admin.conf
```

部署weave网络

```
sysctl net.bridge.bridge-nf-call-iptables=1 -w
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n')"
```

等待coredns pod的状态变成Running，就可以继续添加从节点了
在其它机器上执行完上述操作后，启动work node
```
kubeadm join —token 1111.111111111111 *.*.*.*:6443
```

helm 安装
**helm安装同样也会遇到docker镜像访问不了的问题**

```
docker pull fishead/gcr.io.kubernetes-helm.tiller:v2.12.3
docker tag fishead/gcr.io.kubernetes-helm.tiller:v2.12.3 gcr.io/kubernetes-helm/tiller:v2.12.3
https://storage.googleapis.com/kubernetes-helm/helm-v2.12.3-linux-amd64.tar.gz
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'      
helm init --upgrade --service-account tiller --tiller-image fishead/gcr.io.kubernetes-helm.tiller:v2.12.3
```
