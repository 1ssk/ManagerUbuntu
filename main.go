package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	netio "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo хранит информацию о процессе.
type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
}

// SystemStats содержит общую информацию о системе и список процессов.
type SystemStats struct {
	CPUUsage    float64       `json:"cpu_usage"`
	MemoryTotal float64       `json:"memory_total"` // ГБ
	MemoryUsed  float64       `json:"memory_used"`  // ГБ
	MemoryUsage float64       `json:"memory_usage"`
	DiskTotal   float64       `json:"disk_total"` // ГБ
	DiskUsed    float64       `json:"disk_used"`  // ГБ
	DiskUsage   float64       `json:"disk_usage"`
	NetSent     float64       `json:"net_sent"` // МБ
	NetRecv     float64       `json:"net_recv"` // МБ
	Processes   []ProcessInfo `json:"processes"`
}

// getProcesses получает список всех процессов и сортирует их по убыванию использования CPU.
func getProcesses() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	var result []ProcessInfo
	for _, p := range procs {
		name, _ := p.Name()
		cpuPerc, _ := p.CPUPercent()
		memPerc, _ := p.MemoryPercent()
		result = append(result, ProcessInfo{
			PID:           p.Pid,
			Name:          name,
			CPUPercent:    cpuPerc,
			MemoryPercent: memPerc,
		})
	}
	// Сортировка по убыванию использования CPU
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPUPercent > result[j].CPUPercent
	})
	return result, nil
}

// getSystemStats собирает статистику по системе и список процессов.
func getSystemStats() (SystemStats, error) {
	var stats SystemStats

	// CPU: общая загрузка
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil || len(cpuPercents) == 0 {
		stats.CPUUsage = 0
	} else {
		stats.CPUUsage = cpuPercents[0]
	}

	// Память (конвертируем байты в гигабайты)
	vmem, err := mem.VirtualMemory()
	if err == nil {
		stats.MemoryTotal = float64(vmem.Total) / (1024 * 1024 * 1024)
		stats.MemoryUsed = float64(vmem.Used) / (1024 * 1024 * 1024)
		stats.MemoryUsage = vmem.UsedPercent
	}

	// Диск (корневой раздел, в гигабайтах)
	du, err := disk.Usage("/")
	if err == nil {
		stats.DiskTotal = float64(du.Total) / (1024 * 1024 * 1024)
		stats.DiskUsed = float64(du.Used) / (1024 * 1024 * 1024)
		stats.DiskUsage = du.UsedPercent
	}

	// Сеть (считаем в мегабайтах)
	netIO, err := netio.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		stats.NetSent = float64(netIO[0].BytesSent) / (1024 * 1024)
		stats.NetRecv = float64(netIO[0].BytesRecv) / (1024 * 1024)
	}

	// Процессы
	procs, err := getProcesses()
	if err == nil {
		stats.Processes = procs
	}

	return stats, nil
}

// indexTemplate – HTML-шаблон с JavaScript для обновления данных в реальном времени.
var indexTemplate = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Диспетчер задач Annonyx</title>
	<style>
		@import url('https://fonts.googleapis.com/css2?family=Orbitron:wght@400;700&display=swap');
		body {
			background: #0b0c10;
			color: #66fcf1;
			font-family: 'Orbitron', sans-serif;
			margin: 0;
			padding: 20px;
		}
		h1, h2 {
			color: #45a29e;
		}
		.section {
			margin-bottom: 30px;
			padding: 20px;
			background: #1f2833;
			border-radius: 10px;
			box-shadow: 0 0 10px rgba(0, 255, 255, 0.3);
		}
		table {
			width: 100%;
			border-collapse: collapse;
			margin-top: 15px;
		}
		th, td {
			padding: 12px;
			border: 1px solid #45a29e;
			text-align: center;
		}
		th {
			background: #45a29e;
			color: #0b0c10;
		}
		input[type="submit"] {
			background: #66fcf1;
			border: none;
			color: #0b0c10;
			padding: 8px 16px;
			cursor: pointer;
			font-weight: bold;
			border-radius: 5px;
		}
		input[type="submit"]:hover {
			background: #45a29e;
		}
		p {
			font-size: 1.1em;
		}
	</style>
</head>
<body>
	<h1>Диспетчер задач Annonyx</h1>

	<div class="section" id="system-info">
		<h2>Общая информация о системе</h2>
		<p><strong>Загрузка CPU:</strong> <span id="cpu_usage"></span>%</p>
		<p>
			<strong>Оперативная память:</strong> Всего <span id="mem_total"></span> ГБ, Используется <span id="mem_used"></span> ГБ (<span id="mem_usage"></span>%)
		</p>
		<p>
			<strong>Жёсткий диск (/):</strong> Всего <span id="disk_total"></span> ГБ, Используется <span id="disk_used"></span> ГБ (<span id="disk_usage"></span>%)
		</p>
		<p>
			<strong>Сетевая активность:</strong> Отправлено <span id="net_sent"></span> МБ, Получено <span id="net_recv"></span> МБ
		</p>
	</div>

	<div class="section">
		<h2>Список процессов (отсортировано по загрузке CPU)</h2>
		<table id="processes-table">
			<tr>
				<th>PID</th>
				<th>Название</th>
				<th>Загрузка CPU (%)</th>
				<th>Использование памяти (%)</th>
				<th>Действие</th>
			</tr>
		</table>
	</div>

	<script>
		// Функция для обновления данных с сервера
		async function updateStats() {
			try {
				const response = await fetch('/api/stats');
				const data = await response.json();

				// Обновляем системную информацию
				document.getElementById('cpu_usage').textContent = data.cpu_usage.toFixed(2);
				document.getElementById('mem_total').textContent = data.memory_total.toFixed(2);
				document.getElementById('mem_used').textContent = data.memory_used.toFixed(2);
				document.getElementById('mem_usage').textContent = data.memory_usage.toFixed(2);
				document.getElementById('disk_total').textContent = data.disk_total.toFixed(2);
				document.getElementById('disk_used').textContent = data.disk_used.toFixed(2);
				document.getElementById('disk_usage').textContent = data.disk_usage.toFixed(2);
				document.getElementById('net_sent').textContent = data.net_sent.toFixed(2);
				document.getElementById('net_recv').textContent = data.net_recv.toFixed(2);

				// Обновляем таблицу процессов
				var table = document.getElementById('processes-table');
				// Очищаем таблицу, оставляем только заголовок
				table.innerHTML = "<tr><th>PID</th><th>Название</th><th>Загрузка CPU (%)</th><th>Использование памяти (%)</th><th>Действие</th></tr>";
				data.processes.forEach(function(proc) {
					var row = document.createElement('tr');
					row.innerHTML = "<td>" + proc.pid + "</td>" +
					                "<td>" + proc.name + "</td>" +
					                "<td>" + proc.cpu_percent.toFixed(2) + "</td>" +
					                "<td>" + proc.memory_percent.toFixed(2) + "</td>" +
					                "<td><form method=\"POST\" action=\"/kill\">" +
					                "<input type=\"hidden\" name=\"pid\" value=\"" + proc.pid + "\">" +
					                "<input type=\"submit\" value=\"Завершить задачу\">" +
					                "</form></td>";
					table.appendChild(row);
				});
			} catch (err) {
				console.error('Ошибка обновления данных:', err);
			}
		}

		// Обновляем данные сразу и затем каждые 3 секунды
		updateStats();
		setInterval(updateStats, 3000);
	</script>
</body>
</html>
`))

// indexHandler отдает HTML-страницу.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := indexTemplate.Execute(w, nil)
	if err != nil {
		http.Error(w, "Ошибка отображения шаблона", http.StatusInternalServerError)
	}
}

// statsHandler возвращает данные системы в формате JSON.
func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := getSystemStats()
	if err != nil {
		http.Error(w, "Ошибка получения информации о системе", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// killHandler завершает процесс по PID.
func killHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	pidStr := r.FormValue("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		http.Error(w, "Неверный PID", http.StatusBadRequest)
		return
	}
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		http.Error(w, "Процесс не найден", http.StatusNotFound)
		return
	}
	err = proc.Kill()
	if err != nil {
		http.Error(w, fmt.Sprintf("Не удалось завершить процесс: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/stats", statsHandler)
	http.HandleFunc("/kill", killHandler)
	fmt.Println("Сервер запущен: http://localhost:7000")
	log.Fatal(http.ListenAndServe(":7000", nil))
}
