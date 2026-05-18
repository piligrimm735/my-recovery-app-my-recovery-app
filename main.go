package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	mp4Magic  = []byte{0x66, 0x74, 0x79, 0x70}
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Media Recovery")
	myWindow.Resize(fyne.NewSize(400, 600))

	// Элементы интерфейса
	statusLabel := widget.NewLabel("Нажмите кнопку для начала сканирования")
	logArea := widget.NewMultiLineEntry()
	logArea.SetText("Логи появятся здесь...\n")
	
	startButton := widget.NewButton("Запустить восстановление", func() {
		statusLabel.SetText("Сканирование запущено...")
		logArea.SetText("")

		// Запускаем сканирование в отдельном потоке (горутине), чтобы экран не завис
		go func() {
			rootDir := "/sdcard"
			outputDir := "/sdcard/Восстановленные_Фото"

			os.MkdirAll(outputDir, os.ModePerm)

			restored := 0
			scanned := 0

			filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					if path == outputDir || strings.Contains(path, "Android/data") {
						return filepath.SkipDir
					}
					return nil
				}

				scanned++
				lowerPath := strings.ToLower(path)
				
				if strings.Contains(lowerPath, "thumb") || strings.Contains(lowerPath, "cache") || strings.Contains(lowerPath, "telegram") {
					ext := getRealExt(path)
					if ext != "" {
						restored++
						uniqueName := fmt.Sprintf("restored_%d.%s", time.Now().UnixNano(), ext)
						dest := filepath.Join(outputDir, uniqueName)

						if err := copyFile(path, dest); err == nil {
							// Выводим лог на экран приложения
							logArea.Append(fmt.Sprintf("[+] Найдено: %s\n", uniqueName))
						}
					}
				}
				return nil
			})

			statusLabel.SetText(fmt.Sprintf("Готово! Восстановлено: %d", restored))
		}()
	})

	// Собираем интерфейс в кучу
	scrollLogs := container.NewVScroll(logArea)
	scrollLogs.SetMinSize(fyne.NewSize(380, 400))

	content := container.NewVBox(
		statusLabel,
		startButton,
		scrollLogs,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func getRealExt(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	buf := make([]byte, 16)
	n, err := file.Read(buf)
	if err != nil || n < 4 {
		return ""
	}
	if bytes.HasPrefix(buf, jpegMagic) { return "jpg" }
	if bytes.HasPrefix(buf, pngMagic) { return "png" }
	if bytes.Contains(buf[:n], mp4Magic) { return "mp4" }
	return ""
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil { return err }
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil { return err }
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}