# gvm - Generic Version Manager

![Lisensi](https://img.shields.io/badge/License-MIT-green.svg)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey.svg)
![Go Version](https://img.shields.io/badge/Go-1.18%2B-blue.svg)

**gvm (Generic Version Manager)** adalah manajer versi multi-bahasa berbasis *command-line* yang terinspirasi oleh `nvm`. Dibuat dengan Go, `gvm` memungkinkan Anda untuk mencari, menginstal, dan mengelola beberapa versi dari berbagai bahasa pemrograman dengan mudah dalam satu alat.


Contoh alur kerja di terminal PowerShell:

```powershell
# Mencari versi Python yang tersedia secara online
> gvm search python
Mengambil daftar versi dari [https://www.python.org/ftp/python/](https://www.python.org/ftp/python/)...
Versi python yang tersedia (dari yang terbaru):
3.12.4
3.11.9
3.10.14
...

# Menginstal versi yang diinginkan
> gvm install python 3.11.9
Mengunduh python versi 3.11.9 dari [https://www.python.org/ftp/python/3.11.9/python-3.11.9-embed-amd64.zip](https://www.python.org/ftp/python/3.11.9/python-3.11.9-embed-amd64.zip)
[====================================] 100%
Unduhan selesai. Mengekstrak...
✅ Berhasil menginstal python versi 3.11.9

# Melihat daftar versi yang sudah terinstal
> gvm list python
Versi python yang terinstal:
- 3.11.9

# Menggunakan versi yang baru diinstal untuk sesi terminal ini
> gvm use python 3.11.9 | iex

# Memverifikasi versi aktif
> python --version
Python 3.11.9