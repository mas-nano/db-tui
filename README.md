# db-tui

TUI database client untuk SQLite, MariaDB/MySQL, dan PostgreSQL. Dibangun dengan Bubble Tea v2.

## Prasyarat

- Go 1.24+ (teruji di 1.27.1)

## Jalan cepat

```bash
cd db-tui
go mod tidy
go run .
```

Navigasi: `1` connections, `2` browser, `3` query, `↑/↓` atau `j/k`, `q` quit.

## Config

Salin contoh lalu edit:

```bash
mkdir -p ~/.config/db-tui
cp internal/config/connections.yaml.example ~/.config/db-tui/connections.yaml
```

## Struktur

```text
db-tui/
├── main.go
├── internal/
│   ├── app/            # root Model/Update/View Bubble Tea
│   ├── config/         # load connections.yaml
│   ├── db/             # interface DB + sqlite/postgres/mariadb
│   └── ui/
│       ├── browser/
│       ├── connection/
│       ├── query/
│       └── styles/
```

## Roadmap

- [ ] connect beneran + list tabel
- [ ] query editor (textarea) + result table
- [ ] form tambah koneksi
- [ ] simpan history query
