# gedis
golang版本redis

v0.1 仅支持ping命令

```shell
./redis-cli -h 127.0.0.1 -p 9000
127.0.0.1:9000> ping
pong
```

v0.2 新增select, get, set

```shell
redis-cli.exe -h localhost -p 9000
localhost:9000> select 2
ok
localhost:9000[2]> get a
(nil)
localhost:9000[2]> set a b
ok
localhost:9000[2]> get a
"b"
localhost:9000[2]>
```

v0.3 
1. 新增mset, mget