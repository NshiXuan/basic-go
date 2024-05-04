FROM ubuntu:20.04

# 把编译后的 webook 打包进这个镜像, 放到工作目录 /app 中, app 可以随便换
COPY webook /app/webook
WORKDIR /app

# 执行入口
ENTRYPOINT ["/app/webook"]