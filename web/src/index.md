---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "FoolVPN"
  tagline: The Fool on Service

features:
  - title: Cepat 🔥
    details: Sing-Box + Caddy. Wuzzhhh!!!
  - title: Simpel ⛄️
    details: Kamu bisa akses hal yang kamu perlukan lewat bot dan API
  - title: Fleksibel 🍚
    details: BFR, SFA, V2rayNG ? Semua kamu bisa pake. Support kok!
---

---

|    Key    | Penjelasan                             | Multiple | Contoh              |
| :-------: | :------------------------------------- | :------: | :------------------ |
|   pass    | Masukkan password                      |    No    | &pass=fool          |
|   free    | Pilih VPN gratis saja                  |    No    | &free=1             |
|  premium  | Pilih VPN premium saja                 |    no    | &premium=1          |
|    vpn    | Pilih VPN bersadarkan protokol         |   Yes    | &vpn=vmess          |
|  format   | Pilih format                           |    No    | &format=clash       |
|  region   | Pilih VPN berdasarkan wilayah          |   Yes    | &region=Asia        |
|    cc     | Pilih VPN berdasarkan kode negara      |   Yes    | &cc=SG              |
|  include  | Pilih VPN berdasarkan nama             |   Yes    | &include=melbi      |
|  exclude  | Kecualikan VPN berdasarkan nama        |   Yes    | &exclude=amazon     |
|    tls    | Pilih VPN TLS atau NTLS                |    No    | &tls=1              |
|  network  | Pilih VPN berdasarkan tipe jaringan    |   Yes    | &network=ws         |
|   port    | Pilih VPN berdasarkan port             |   Yes    | &port=80,443        |
|    ip     | Merubah domain host VPN menjadi IP     |    No    | &ip=1               |
|    sni    | Isi VPN dengan host SNI                |   Yes    | &sni=sni.host       |
|    cdn    | Isi VPN dengan host CDN                |   Yes    | &cdn=cdn.host       |
|   mode    | Pilih mode koneksi                     |   Yes    | &mode=cdn           |
|   limit   | Batasi jumlah VPN                      |    No    | &limit=10           |
| subdomain | Menambahkan subdomain ke domain server |    no    | &subdomain=zoom.us  |
|    arg    | Masukkan pair key-value                |   Yes    | &arg=xudp,key:value |

Key yang mendukung input lebih dari 1 (Multiple), dipisahkan menggunakan tanda koma (,)  
Contoh: [Get Free Account](http://fool.azurewebsites.net/get?format=raw&cdn=104.18.2.2&sni=google.com&mode=cdn,sni&region=Asia&vpn=vmess,vless,trojan&pass=megalodon)
