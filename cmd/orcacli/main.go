package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"

	"orca/pkg/container"

	"github.com/spf13/cobra"
)

const (
	defaultServerURL = "http://localhost:8080"
	orcaBanner = `
 ██████╗ ██████╗  ██████╗ █████╗ 
██╔═══██╗██╔══██╗██╔════╝██╔══██╗
██║   ██║██████╔╝██║     ███████║
██║   ██║██╔══██╗██║     ██╔══██║
╚██████╔╝██║  ██║╚██████╗██║  ██║
 ╚═════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
                                  
Container Orchestrator CLI v1.0.0
`
)

var (
	serverURL string
	rootCmd   = &cobra.Command{
		Use:   "orca",
		Short: "🐋 ORCA Container Orchestrator CLI",
		Long: orcaBanner + `
ORCA, Docker konteynerlerini, deployment'ları ve servisleri yönetmek için 
geliştirilmiş basit ve güçlü bir konteyner orkestratörüdür.

Kullanım örnekleri:
  orca create examples/test-container.json  # Yeni konteyner oluştur
  orca start my-container                   # Konteyneri başlat
  orca containers                           # Tüm konteynerleri listele
  orca stats                                # Sistem istatistiklerini göster

Daha fazla bilgi için: orca [komut] --help`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Banner'ı sadece help ve version dışındaki komutlarda göster
			if cmd.Name() != "help" && cmd.Name() != "version" && !cmd.HasParent() {
				fmt.Print(orcaBanner)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", defaultServerURL, "ORCA sunucu URL'si")

	// Container commands
	rootCmd.AddCommand(createContainerCmd)
	rootCmd.AddCommand(listContainersCmd)
	rootCmd.AddCommand(startContainerCmd)
	rootCmd.AddCommand(stopContainerCmd)
	rootCmd.AddCommand(removeContainerCmd)
	rootCmd.AddCommand(inspectContainerCmd)
	rootCmd.AddCommand(logsContainerCmd)

	// Deployment commands
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(listDeploymentsCmd)
	rootCmd.AddCommand(deleteDeploymentCmd)

	// Service commands
	rootCmd.AddCommand(createServiceCmd)
	rootCmd.AddCommand(listServicesCmd)
	rootCmd.AddCommand(deleteServiceCmd)

	// Utility commands
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("❌ Hata: %v\n", err)
		os.Exit(1)
	}
}

// Container commands
var createContainerCmd = &cobra.Command{
	Use:   "create [spec-file]",
	Short: "📦 Yeni bir konteyner oluştur",
	Long: `Belirtilen JSON spec dosyasından yeni bir konteyner oluşturur.

Örnek kullanım:
  orca create examples/test-container.json
  orca create my-app-spec.json`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile := args[0]
		
		fmt.Printf("📄 Spec dosyası okunuyor: %s\n", specFile)
		data, err := ioutil.ReadFile(specFile)
		if err != nil {
			fmt.Printf("❌ Spec dosyası okunamadı: %v\n", err)
			os.Exit(1)
		}

		var spec container.ContainerSpec
		if err := json.Unmarshal(data, &spec); err != nil {
			fmt.Printf("❌ Spec dosyası parse edilemedi: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🚀 Konteyner oluşturuluyor: %s\n", spec.Name)
		c, err := createContainer(spec)
		if err != nil {
			fmt.Printf("❌ Konteyner oluşturulamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Konteyner başarıyla oluşturuldu!\n")
		fmt.Printf("   📋 ID: %s\n", c.ID)
		fmt.Printf("   🏷️  İsim: %s\n", c.Name)
		fmt.Printf("   🖼️  Image: %s\n", c.Image)
		fmt.Printf("   📊 Durum: %s\n", c.Status)
	},
}

var listContainersCmd = &cobra.Command{
	Use:     "containers",
	Aliases: []string{"ps", "list"},
	Short:   "📋 Konteynerleri listele",
	Long: `Sistemdeki tüm konteynerleri durumlarıyla birlikte listeler.

Örnek kullanım:
  orca containers
  orca ps
  orca list`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Konteynerler getiriliyor...")
		containers, err := listContainers()
		if err != nil {
			fmt.Printf("❌ Konteyner listesi alınamadı: %v\n", err)
			os.Exit(1)
		}

		if len(containers) == 0 {
			fmt.Println("📭 Hiç konteyner bulunamadı.")
			return
		}

		fmt.Printf("\n📦 Toplam %d konteyner bulundu:\n\n", len(containers))
		
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tİSİM\tIMAGE\tDURUM\tPORTLAR")
		fmt.Fprintln(w, strings.Repeat("─", 80))
		
		for _, c := range containers {
			ports := ""
			if len(c.Ports) > 0 {
				portStrs := make([]string, 0, len(c.Ports))
				for containerPort, hostPort := range c.Ports {
					portStrs = append(portStrs, fmt.Sprintf("%s:%s", hostPort, containerPort))
				}
				ports = strings.Join(portStrs, ", ")
			}
			
			status := c.Status
			switch status {
			case "running":
				status = "🟢 " + status
			case "exited":
				status = "🔴 " + status
			case "created":
				status = "🟡 " + status
			default:
				status = "⚪ " + status
			}
			
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
				c.ID[:12], c.Name, c.Image, status, ports)
		}
		w.Flush()
	},
}

var startContainerCmd = &cobra.Command{
	Use:   "start [container-name]",
	Short: "🚀 Konteyneri başlat",
	Long: `Belirtilen konteyner adı veya ID'si ile konteyneri başlatır.

Örnek kullanım:
  orca start my-container
  orca start test-integration`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerID := args[0]
		
		fmt.Printf("🚀 Konteyner başlatılıyor: %s\n", containerID)
		if err := startContainer(containerID); err != nil {
			fmt.Printf("❌ Konteyner başlatılamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Konteyner başarıyla başlatıldı: %s\n", containerID)
	},
}

var stopContainerCmd = &cobra.Command{
	Use:   "stop [container-name]",
	Short: "⏹️  Konteyneri durdur",
	Long: `Belirtilen konteyner adı veya ID'si ile konteyneri durdurur.

Örnek kullanım:
  orca stop my-container
  orca stop test-integration`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerID := args[0]
		
		fmt.Printf("⏹️  Konteyner durduruluyor: %s\n", containerID)
		if err := stopContainer(containerID); err != nil {
			fmt.Printf("❌ Konteyner durdurulamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Konteyner başarıyla durduruldu: %s\n", containerID)
	},
}

var removeContainerCmd = &cobra.Command{
	Use:   "remove [container-name]",
	Aliases: []string{"rm", "delete"},
	Short: "🗑️  Konteyneri sil",
	Long: `Belirtilen konteyner adı veya ID'si ile konteyneri sistemden siler.

Örnek kullanım:
  orca remove my-container
  orca rm test-integration
  orca delete old-container`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerID := args[0]
		
		fmt.Printf("🗑️  Konteyner siliniyor: %s\n", containerID)
		if err := removeContainer(containerID); err != nil {
			fmt.Printf("❌ Konteyner silinemedi: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Konteyner başarıyla silindi: %s\n", containerID)
	},
}

var inspectContainerCmd = &cobra.Command{
	Use:   "inspect [container-name]",
	Short: "🔍 Konteyner detaylarını görüntüle",
	Long: `Belirtilen konteyner adı veya ID'si ile konteyner hakkında detaylı bilgi görüntüler.

Örnek kullanım:
  orca inspect my-container
  orca inspect test-integration`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerID := args[0]
		
		fmt.Printf("🔍 Konteyner bilgileri getiriliyor: %s\n", containerID)
		c, err := inspectContainer(containerID)
		if err != nil {
			fmt.Printf("❌ Konteyner bilgileri alınamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n📋 Konteyner Detayları:\n")
		fmt.Printf("═══════════════════════════════════════\n")
		fmt.Printf("🏷️  İsim: %s\n", c.Name)
		fmt.Printf("📋 ID: %s\n", c.ID)
		fmt.Printf("🖼️  Image: %s\n", c.Image)
		fmt.Printf("📊 Durum: %s\n", c.Status)
		
		if len(c.Ports) > 0 {
			fmt.Printf("🌐 Portlar:\n")
			for containerPort, hostPort := range c.Ports {
				fmt.Printf("   %s:%s (tcp)\n", hostPort, containerPort)
			}
		}
		
		if len(c.Environment) > 0 {
			fmt.Printf("🔧 Ortam Değişkenleri:\n")
			for _, env := range c.Environment {
				fmt.Printf("   %s\n", env)
			}
		}
		
		fmt.Printf("📅 Oluşturulma: %s\n", c.Created.Format("2006-01-02 15:04:05"))
		if !c.Started.IsZero() {
			fmt.Printf("🚀 Başlatılma: %s\n", c.Started.Format("2006-01-02 15:04:05"))
		}
	},
}

var logsContainerCmd = &cobra.Command{
	Use:   "logs [container-name]",
	Short: "📜 Konteyner loglarını görüntüle",
	Long: `Belirtilen konteyner adı veya ID'si ile konteyner loglarını görüntüler.

Örnek kullanım:
  orca logs my-container
  orca logs test-integration --tail 50`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerID := args[0]
		tail, _ := cmd.Flags().GetInt("tail")
		
		fmt.Printf("📜 Konteyner logları getiriliyor: %s (son %d satır)\n", containerID, tail)
		logs, err := getContainerLogs(containerID, tail)
		if err != nil {
			fmt.Printf("❌ Konteyner logları alınamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n📋 Konteyner Logları:\n")
		fmt.Printf("═══════════════════════════════════════\n")
		fmt.Print(logs)
	},
}

// Deployment commands
var deployCmd = &cobra.Command{
	Use:   "deploy [spec-file]",
	Short: "Create a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile := args[0]
		
		data, err := ioutil.ReadFile(specFile)
		if err != nil {
			fmt.Printf("Spec dosyası okunamadı: %v\n", err)
			os.Exit(1)
		}

		var spec container.DeploymentSpec
		if err := json.Unmarshal(data, &spec); err != nil {
			fmt.Printf("Spec dosyası parse edilemedi: %v\n", err)
			os.Exit(1)
		}

		deployment, err := createDeployment(spec)
		if err != nil {
			fmt.Printf("Deployment oluşturulamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Deployment oluşturuldu: %s (%d replicas)\n", deployment.Name, len(deployment.Replicas))
	},
}

var listDeploymentsCmd = &cobra.Command{
	Use:     "deployments",
	Aliases: []string{"deploy"},
	Short:   "List deployments",
	Run: func(cmd *cobra.Command, args []string) {
		deployments, err := listDeployments()
		if err != nil {
			fmt.Printf("Deployment listesi alınamadı: %v\n", err)
			os.Exit(1)
		}

		if len(deployments) == 0 {
			fmt.Println("Hiç deployment bulunamadı.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tREPLICAS\tSTATUS\tCREATED")
		
		for _, d := range deployments {
			created := d.Created.Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", 
				d.Name, len(d.Replicas), d.Status, created)
		}
		
		w.Flush()
	},
}

var deleteDeploymentCmd = &cobra.Command{
	Use:   "delete-deployment [name]",
	Short: "Delete a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		
		if err := deleteDeployment(name); err != nil {
			fmt.Printf("Deployment silinemedi: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Deployment silindi: %s\n", name)
	},
}

// Service commands
var createServiceCmd = &cobra.Command{
	Use:   "create-service [spec-file]",
	Short: "Create a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specFile := args[0]
		
		data, err := ioutil.ReadFile(specFile)
		if err != nil {
			fmt.Printf("Spec dosyası okunamadı: %v\n", err)
			os.Exit(1)
		}

		var spec container.ServiceSpec
		if err := json.Unmarshal(data, &spec); err != nil {
			fmt.Printf("Spec dosyası parse edilemedi: %v\n", err)
			os.Exit(1)
		}

		service, err := createService(spec)
		if err != nil {
			fmt.Printf("Service oluşturulamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Service oluşturuldu: %s\n", service.Name)
	},
}

var listServicesCmd = &cobra.Command{
	Use:     "services",
	Aliases: []string{"svc"},
	Short:   "List services",
	Run: func(cmd *cobra.Command, args []string) {
		services, err := listServices()
		if err != nil {
			fmt.Printf("Service listesi alınamadı: %v\n", err)
			os.Exit(1)
		}

		if len(services) == 0 {
			fmt.Println("Hiç service bulunamadı.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tPORTS\tSTATUS\tCREATED")
		
		for _, s := range services {
			ports := formatServicePorts(s.Spec.Ports)
			created := s.Created.Format("2006-01-02 15:04:05")
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
				s.Name, s.Spec.Type, ports, s.Status, created)
		}
		
		w.Flush()
	},
}

var deleteServiceCmd = &cobra.Command{
	Use:   "delete-service [name]",
	Short: "Delete a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		
		if err := deleteService(name); err != nil {
			fmt.Printf("Service silinemedi: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Service silindi: %s\n", name)
	},
}

// Utility commands
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "📊 Sistem istatistiklerini göster",
	Long: `ORCA sisteminin genel durumu ve istatistiklerini görüntüler.

Örnek kullanım:
  orca stats`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 Sistem istatistikleri getiriliyor...")
		stats, err := getStats()
		if err != nil {
			fmt.Printf("❌ İstatistikler alınamadı: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n🐋 ORCA Sistem İstatistikleri:\n")
		fmt.Printf("═══════════════════════════════════════\n")
		
		// Containers bilgisini güvenli şekilde al
		if containers, ok := stats["containers"].(map[string]interface{}); ok {
			total := int(containers["total"].(float64))
			running := int(containers["running"].(float64))
			fmt.Printf("📦 Konteynerler: %d toplam, %d çalışıyor\n", total, running)
			
			if total > 0 {
				stopped := total - running
				fmt.Printf("   🟢 Çalışan: %d\n", running)
				fmt.Printf("   🔴 Durmuş: %d\n", stopped)
			}
		} else {
			fmt.Printf("📦 Konteynerler: 0 toplam, 0 çalışıyor\n")
		}
		
		// Deployments ve Services bilgisini güvenli şekilde al
		if deployments, ok := stats["deployments"].(float64); ok {
			fmt.Printf("🚀 Deployment'lar: %d\n", int(deployments))
		} else {
			fmt.Printf("🚀 Deployment'lar: 0\n")
		}
		
		if services, ok := stats["services"].(float64); ok {
			fmt.Printf("🌐 Servisler: %d\n", int(services))
		} else {
			fmt.Printf("🌐 Servisler: 0\n")
		}
		
		fmt.Printf("\n✅ Sistem sağlıklı ve çalışıyor!\n")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "ℹ️  Sürüm bilgilerini göster",
	Long: `ORCA CLI ve sistem sürüm bilgilerini görüntüler.

Örnek kullanım:
  orca version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(orcaBanner)
		fmt.Printf("\n📋 Sürüm Bilgileri:\n")
		fmt.Printf("═══════════════════════════════════════\n")
		fmt.Printf("🐋 ORCA CLI: v1.0.0\n")
		fmt.Printf("🔧 Go Runtime: %s\n", "go1.21+")
		fmt.Printf("🏗️  Build: Production\n")
		fmt.Printf("📅 Release Date: 2025-01-28\n")
		fmt.Printf("\n💡 Daha fazla bilgi için: orca --help\n")
	},
}

func init() {
	logsContainerCmd.Flags().Int("tail", 100, "Number of lines to show from the end of the logs")
}