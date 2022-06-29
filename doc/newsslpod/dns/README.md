通过 context 携带数据进行 dns 查询：

1. dns_server string 通过指定 dns 服务器进行查询
2. dns_aaaa bool 是否查询 AAAA，这将查询 A + AAAA
3. dns_eaddr string 通过指定地址查询 EDNS，这会在支持 EDNS 的 DNS 服务器中查询。
4. gfw.InGfwContext(ctx) 真，将从香港 DNS（852）服务器进行查询。
