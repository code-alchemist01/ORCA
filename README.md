# Orca Container Orchestrator

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Platform](https://img.shields.io/badge/Platform-Windows-0078D4?style=flat&logo=windows)](https://www.microsoft.com/windows/)

Orca, Docker konteynerlerini, deployment'ları ve servisleri yönetmek için geliştirilmiş basit ve güçlü bir konteyner orkestratörüdür. RESTful API ve kullanıcı dostu CLI arayüzü ile container lifecycle management'ı kolaylaştırır.

## Özellikler

- **Container Yönetimi**: Docker konteynerlerini oluşturma, başlatma, durdurma, silme ve izleme
- **Deployment Yönetimi**: Uygulama deployment'larını oluşturma, listeleme ve yönetme
- **Service Yönetimi**: Servisleri oluşturma, listeleme ve yönetme
- **Container İnceleme**: Container detaylarını görüntüleme ve log takibi
- **RESTful API**: HTTP API üzerinden tüm işlemleri gerçekleştirme
- **CLI Arayüzü**: Komut satırı üzerinden kolay kullanım
- **Sistem İzleme**: Gerçek zamanlı sistem istatistikleri
- **Port Yönetimi**: Container port mapping ve service discovery
- **Logging**: Yapılandırılabilir loglama sistemi
- **Persistent Storage**: Deployment ve servis bilgilerinin kalıcı saklanması

## Kurulum

### Gereksinimler

- Go 1.21 veya üzeri
- Docker Desktop (Windows için)
- Git

### Projeyi İndirme

```bash
# Repository'yi clone edin
git clone https://github.com/yourusername/orca-container-orchestrator.git
cd orca-container-orchestrator
```

### Build

Projeyi build etmek için:

```bash
# Tüm bağımlılıkları indir
go mod tidy

# Orchestrator ve CLI'ı build et
make build

# Veya Windows için
build.bat
```

### Kurulum

Windows için CLI'ı sistem genelinde kullanmak için:

```bash
# Yönetici olarak çalıştır
install.bat
```

## Kullanım

### Orchestrator Sunucusunu Başlatma

```bash
# Orchestrator'ı başlat
.\bin\orchestrator.exe

# Veya make ile
make run
```

Sunucu varsayılan olarak `localhost:8080` adresinde çalışır.

### CLI Kullanımı

```bash
# Yardım menüsü
.\bin\orca.exe --help

# Container listesi
.\bin\orca.exe containers

# Container oluşturma
.\bin\orca.exe create examples/container-spec.json

# Container başlatma
.\bin\orca.exe start <container-name>

# Container durdurma
.\bin\orca.exe stop <container-name>

# Container logları
.\bin\orca.exe logs <container-name>

# Container detayları
.\bin\orca.exe inspect <container-name>

# Deployment oluşturma
.\bin\orca.exe deploy examples/deployment-spec.json

# Deployment listesi
.\bin\orca.exe deployments

# Service oluşturma
.\bin\orca.exe create-service examples/service-spec.json

# Service listesi
.\bin\orca.exe services

# Sistem istatistikleri
.\bin\orca.exe stats
```

## Konfigürasyon

Konfigürasyon dosyası `config/config.yaml` konumunda bulunur:

```yaml
server:
  host: "localhost"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

docker:
  host: "unix:///var/run/docker.sock"  # Linux/macOS
  # host: "npipe:////./pipe/docker_engine"  # Windows
  version: "1.24"

storage:
  data_dir: "./data"
  backup_enabled: true
  backup_interval: "1h"

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

## Örnek Dosyalar

`examples/` klasöründe örnek spec dosyaları bulunur:

- `container-spec.json`: Container oluşturma örneği
- `deployment-spec.json`: Deployment oluşturma örneği
- `service-spec.json`: Service oluşturma örneği

## API Endpoints

### Container Endpoints

- `GET /containers` - Container listesi
- `POST /containers` - Container oluştur
- `GET /containers/{name}` - Container detayı
- `POST /containers/{name}/start` - Container başlat
- `POST /containers/{name}/stop` - Container durdur
- `DELETE /containers/{name}` - Container sil
- `GET /containers/{name}/logs` - Container logları

### Deployment Endpoints

- `GET /deployments` - Deployment listesi
- `POST /deployments` - Deployment oluştur
- `GET /deployments/{name}` - Deployment detayı
- `DELETE /deployments/{name}` - Deployment sil

### Service Endpoints

- `GET /services` - Service listesi
- `POST /services` - Service oluştur
- `GET /services/{name}` - Service detayı
- `DELETE /services/{name}` - Service sil

### Diğer Endpoints

- `GET /health` - Health check
- `GET /stats` - Sistem istatistikleri

## Geliştirme

### Proje Yapısı

```
├── cmd/
│   ├── orchestrator/    # Orchestrator uygulaması
│   └── orcacli/        # CLI uygulaması
├── pkg/
│   ├── config/         # Konfigürasyon yönetimi
│   ├── container/      # Container yönetimi
│   ├── scheduler/      # Deployment ve service yönetimi
│   └── storage/        # Veri saklama
├── config/             # Konfigürasyon dosyaları
├── examples/           # Örnek spec dosyaları
└── bin/               # Build edilmiş executable'lar
```

### Test

```bash
# Tüm testleri çalıştır
go test ./...

# Veya make ile
make test
```

### Linting ve Formatting

```bash
# Kodu formatla
make fmt

# Lint kontrolü
make lint
```

## Docker Daemon Gereksinimi

Orca, Docker API'sini kullanarak konteynerları yönetir. Bu nedenle:

- **Windows**: Docker Desktop'ın çalışıyor olması gerekir
- **Linux/macOS**: Docker daemon'ın çalışıyor olması gerekir

Docker daemon çalışmıyorsa, container işlemleri başarısız olur.

## Sorun Giderme

### Docker Bağlantı Hatası

```
error during connect: this error may indicate that the docker daemon is not running
```

**Çözüm**: Docker Desktop'ı başlatın veya Docker daemon'ın çalıştığından emin olun.

### Port Kullanımda Hatası

```
bind: address already in use
```

**Çözüm**: 8080 portunu kullanan başka bir uygulamayı durdurun veya config dosyasından portu değiştirin.

### HTTP Status Code Sorunları

Eğer deployment veya service oluşturma sırasında başarılı yanıt almasına rağmen "failed" hatası alıyorsanız:

```
Deployment created successfully but CLI shows failure
```

**Çözüm**: Bu sorun v1.0.0'da düzeltilmiştir. CLI artık hem HTTP 200 hem de HTTP 201 kodlarını başarılı olarak kabul eder.

### Mevcut Deployment/Service Hatası

```
HTTP 500: deployment/service already exists
```

**Çözüm**: Önce mevcut deployment veya service'i silin:
```bash
# HTTP DELETE ile silme
Invoke-WebRequest -Uri "http://localhost:8080/deployments/<name>" -Method DELETE
Invoke-WebRequest -Uri "http://localhost:8080/services/<name>" -Method DELETE
```

## Lisans

Bu proje [MIT lisansı](LICENSE) altında lisanslanmıştır. Detaylar için `LICENSE` dosyasına bakınız.

## Katkıda Bulunma

Katkılarınızı memnuniyetle karşılıyoruz! Katkıda bulunmak için:

1. Bu repository'yi fork edin
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Değişikliklerinizi commit edin (`git commit -m 'Add some amazing feature'`)
4. Branch'inizi push edin (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

### Geliştirme Kuralları

- Go kod standartlarına uyun (`go fmt`, `go vet`)
- Yeni özellikler için testler yazın
- Commit mesajlarını açıklayıcı yazın
- Büyük değişiklikler için önce issue açın

## Versiyon

**v1.0.0** - İlk stabil sürüm (2024)

### Özellikler
- ✅ **Container Orchestration**: Tam fonksiyonel container lifecycle management
- ✅ **Deployment Management**: Multi-replica deployment desteği
- ✅ **Service Discovery**: Port mapping ve networking
- ✅ **RESTful API**: Kapsamlı HTTP API endpoints
- ✅ **CLI Interface**: Kullanıcı dostu komut satırı arayüzü
- ✅ **System Monitoring**: Gerçek zamanlı istatistikler
- ✅ **Error Handling**: Güvenilir hata yönetimi
- ✅ **Logging System**: Yapılandırılabilir log sistemi
- ✅ **Persistent Storage**: JSON tabanlı veri saklama

### Test Durumu
- ✅ Container operations (create, start, stop, remove, inspect, logs)
- ✅ Deployment lifecycle (create, list, delete, multi-replica)
- ✅ Service management (create, list, delete, port mapping)
- ✅ API endpoints (containers, deployments, services, health, stats)
- ✅ CLI-API communication (HTTP status code handling)
- ✅ Error scenarios ve edge cases

## İletişim

- **Issues**: Sorularınız ve bug raporları için [GitHub Issues](../../issues) kullanın
- **Pull Requests**: Katkılarınızı [Pull Requests](../../pulls) ile gönderin
- **Discussions**: Genel tartışmalar için [GitHub Discussions](../../discussions) kullanın

## Teşekkürler

Bu projeyi geliştirirken kullanılan açık kaynak projeler:
- [Docker](https://www.docker.com/) - Container runtime
- [Go](https://golang.org/) - Programming language
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router