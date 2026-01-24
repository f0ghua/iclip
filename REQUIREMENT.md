我的开发环境是在 Imera 里面，通过 VMWare Horizon Client 使用 rdp 登录远程  ubuntu 机器使用。由于 Imera 的限制，从外往内复制黏贴是可以的，但是从 ubuntu 桌面系统往外部拷贝数据会被限制。

目前 Imera 从内往外可以通过 samba 的方式通讯

- 把要发送的内容打包成 tar.gz 压缩文件 xxx.tar.gz
- 使用 @reference/imera_tx2.sh <xxx.tar.gz> 发送
- 然后在 imera 外面通过访问 \172.21.68.235\Receive2\PD-6-4\ 取得这个 xxx.tar.gz

为了简化使用，我希望写一个程序 iclip，用来把内部的 clipboard 内容实时的传输到 imera 外部。

程序应该提供简单清晰的界面，分为 iclip_client 和 iclip_server。

- iclip_client 运行在内部的 ubuntu 机器上，负责监控 clipboard 内容，实时显示。每次有变化，就打包成 clipboard_<用户名>.tar.gz 使用和 imera_tx2.sh 相同的方法丢出去
- iclip_server 运行在外部的 windows pc 上，负责监控 \172.21.68.235\Receive2\PD-6-4\ 里的内容，如果有 clipboard 的文件，则把内容读取出来放到 clipboard 里去，同时显示出来。然后删除那个文件。