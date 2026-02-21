---
description: 'Requirement application'
---

## Project Overview
Aplikasi Management Club Sport untuk perusahaan XYZ — sistem manajemen tim sepak bola yang memungkinkan admin perusahaan mengelola tim, pemain, jadwal pertandingan, pelaporan hasil, dan laporan ringkasan pertandingan.

## Feature
- Manajemen Tim: tambah, ubah, hapus tim; simpan nama, logo, tahun berdiri, alamat, kota.
- Manajemen Pemain: tambah, ubah, hapus pemain; simpan nama, tinggi, berat, posisi, nomor punggung; satu pemain hanya pada satu tim; nomor punggung unik per tim.
- Jadwal Pertandingan: buat dan ubah jadwal; simpan tanggal, waktu, tim home, tim away; validasi tim berbeda.
- Pelaporan Hasil: lapor skor akhir, daftar gol per pemain beserta menit terjadinya; konsistensi antara gol dan skor dijaga.
- Laporan Hasil dan Statistik: tampilkan jadwal, skor, status akhir (Home Win / Away Win / Draw), top scorer per pertandingan, akumulasi kemenangan home dan away sampai pertandingan tersebut.
- Keamanan dan Audit: autentikasi admin, RBAC, dan audit log untuk semua perubahan.

## Model
- Team: id, name, logo_url, founded_year, address, city, created_at, updated_at.
- Player: id, team_id, name, height_cm, weight_kg, position, shirt_number, created_at, updated_at.
- Match: id, match_date, match_time, home_team_id, away_team_id, home_score, away_score, status, created_at, updated_at.
- Goal: id, match_id, player_id, minute_scored, created_at.
- User(Admin) dan AuditLog.
Kunci Validasi:
- UNIQUE(team_id, shirt_number) pada tabel Player.
- CHECK(home_team_id <> away_team_id) pada tabel Match.
- Saat menambah Goal, pastikan player.team_id adalah salah satu tim yang bermain pada match


## API Desilg

Endpoint | Method | Function
/teams | GET | List Team
/teams | POST | Add Team
/teams/{id} | GET | Detail Team
/teams/{id} | PUT | Update Team
/players | GET | list player (can be fillter by team)
/players | POST | Add player
/players{id} | GET | Detail player
/players{id} | PUT | Update player
/matches | GET | list match
/matches | POST | add new schedule match
/matches/{id} | PUT | Update scheduke match
/matches/{id}/goals | POST | add goal score (player,minute)
/matches/{id}/start | PATCH | update status match 
/matches/{id}/end | PATCH | update status match
/report/match-results | GET | report match

## Business Rules
- Satu pemain hanya pada satu tim: Player.team_id wajib; pindah tim dilakukan lewat update team_id.
- Nomor punggung unik per tim: enforce UNIQUE(team_id, shirt_number) di DB; validasi juga di API.
- Tidak boleh home_team == away_team: validasi di API dan constraint DB.
- Menambahkan gol: hanya pemain yang terdaftar di salah satu tim yang bermain pada match boleh dicatat sebagai pencetak gol; validasi cross-check player.team_id ∈ {home_team_id, away_team_id}.
- Pelaporan hasil: saat status diubah ke finished, sistem harus:
  - Menyimpan home_score dan away_score.
  - Menyimpan entri Goal sesuai payload; konsistensi antara jumlah gol dan skor akhir harus dicek.
  - Menulis entry ke AuditLog.
- Waktu pertandingan: simpan sebagai match_date + match_time (timezone-aware jika perlu).
- Hak akses: hanya admin (role) yang dapat CRUD tim, pemain, jadwal, dan melaporkan hasil.
