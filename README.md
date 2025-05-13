# Tubes2_BE_akuAdalahStima2

Repositori ini berisi program back-end tugas besar 2 akuAdalahStima2. Program ini bertugas sebagai endpoint API dan digunakan untuk proses komputasi algortima dfs, bfs, dan scraper.

## Description

Program ini memanfaatkan algoritma DFS dan BFS untuk mencari resep dalam pembuatan suatu elemen pada permainan https://littlealchemy2.com

Algoritma DFS (Depth-First Search) diterapkan secara iteratif untuk menelusuri kombinasi elemen dari sebuah pohon resep. Proses eksplorasi dimulai dari node target, dengan memprioritaskan bagian subtree kiri (left) terlebih dahulu secara paralel, lalu dilanjutkan ke subtree kanan (right).

Untuk menjamin proses backtracking yang aman, setiap node memiliki salinan status visited sendiri menggunakan struktur map yang di-clone. Algoritma ini juga dioptimalkan dengan membatasi jumlah kombinasi subtree kiri yang valid hingga mencapai jumlah resep yang diminta pengguna. Jika batas tercapai, pencarian di subtree kiri dihentikan dan hanya satu kombinasi dari subtree kanan yang diproses untuk menyeimbangkan total hasil. Suatu subtree dianggap valid jika menghasilkan base element (air, earth, fire, water).

Algoritma BFS diterapkan untuk membangun pohon resep secara bertahap berdasarkan kedalaman dari elemen target. Sistem akan mengeksplorasi berbagai kombinasi bahan dengan pendekatan level-by-level, sehingga seluruh node pada level tertentu diselesaikan sebelum lanjut ke level berikutnya. Setiap kombinasi yang valid kemudian disusun menjadi pohon resep hingga batas jumlah yang ditentukan tercapai. Proses ini juga dilengkapi dengan dukungan multithreading untuk meningkatkan performa, serta pencatatan metrik seperti jumlah simpul yang dikunjungi dan durasi eksekusi.

## Prerequisite

Sebelum memulai, pastikan Anda telah menginstal:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

Untuk memastikan Docker telah terinstal, jalankan perintah berikut:

```bash
docker --version
docker compose version
```

## Clone Repository

Clone repository ini ke dalam komputer Anda:

```bash
git clone https://github.com/AlfianHanifFY/Tubes2_BE_akuAdalahStima2.git
cd Tubes2_BE_akuAdalahStima2
```

## Build Docker

```bash
docker compose build
```

## Menjalankan Aplikasi

Jalankan container:

```bash
docker compose up
```

Gunakan flag `-d` untuk menjalankan dalam mode _detached_:

```bash
docker compose up -d
```

Program akan berjalan pada `localhost:8080` dan berfungsi sebagai endpoint API untuk program front-end

## Menghentikan Aplikasi

Untuk menghentikan dan menghapus container:

```bash
docker compose down
```

---

## Struktur Proyek

```text
.
├── BFS
│   └── MultipleRecipeBFS.go
├── DFS
│   └── MultipleRecipeDFS.go
├── Dockerfile
├── Element
│   ├── Element.go
│   └── Tree.go
├── Handler
│   ├── BFSHandler.go
│   ├── DFSHandler.go
│   └── ScrapperHandler.go
├── README.md
├── docker-compose.yml
├── go.mod
├── go.sum
├── main.go
└── scrapper
    └── scrapper.go
```

---

## 📬 Kontributor

| Nama                         | NIM      |
| ---------------------------- | -------- |
| Alfian Hanif Fitria Yustanto | 13523073 |
| Heleni Gratia M. Tampubolon  | 13523107 |
| Ahmad Wicaksono              | 13523121 |
