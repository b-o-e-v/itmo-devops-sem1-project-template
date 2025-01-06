package server

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"project_sem/pkg/db"
	"time"

	"github.com/gin-gonic/gin"
)

// ответ на загрузку архива
type LoadResponse struct {
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}

// функция для получения файла из запроса
func getFileFromRequest(ctx *gin.Context) (*bytes.Buffer, error) {
	// получаем файл из запроса
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("invalid file")
	}
	defer file.Close()

	// копируем содержимое файла в буфер
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, fmt.Errorf("failed to read file")
	}

	return buf, nil
}

// функция для открытия ZIP-архива
func openZipArchive(buf *bytes.Buffer) (*zip.Reader, error) {
	// открываем ZIP-архив из буфера
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return nil, fmt.Errorf("invalid zip file")
	}
	return zipReader, nil
}

// функция для извлечения CSV-файла из архива
func extractCSVFileFromZip(zipReader *zip.Reader) (io.ReadCloser, error) {
	// ищем файл data.csv в архиве
	for _, file := range zipReader.File {
		if file.Name == "data.csv" {
			csvFile, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open CSV file")
			}
			return csvFile, nil
		}
	}

	// если файл data.csv не найден
	return nil, fmt.Errorf("data.csv not found in zip file")
}

// функция для чтения данных из CSV
func readCSVFile(csvFile io.ReadCloser) ([][]string, error) {
	// создаем CSV-ридер для обработки содержимого файла, пропуская заголовок
	reader := csv.NewReader(csvFile)
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to skip header in CSV file")
	}

	// читаем записи из CSV-файла построчно
	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV file")
		}
		if len(record) < 5 {
			return nil, fmt.Errorf("invalid record in CSV file")
		}
		records = append(records, record)
	}
	return records, nil
}

// функция для обработки транзакции и вставки данных в базу
func processDatabaseTransaction(records [][]string) (int, int, float64, error) {
	tx, err := db.Conn.Begin()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback()

	var totalItems int

	// вставляем данные в базу данных
	for _, record := range records {
		id, name, category, price, createdAt := record[0], record[1], record[2], record[3], record[4]

		result, err := tx.Exec(
			"INSERT INTO prices (id, created_at, name, category, price) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING",
			id, createdAt, name, category, price,
		)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("failed to execute query")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return 0, 0, 0, fmt.Errorf("failed to get affected rows")
		}

		if rowsAffected > 0 {
			totalItems++
		}
	}

	// получаем общее количество категорий и цен
	row := tx.QueryRow("SELECT COUNT(DISTINCT category), SUM(price) FROM prices")
	var totalCategoriesDB int
	var totalPriceDB float64
	if err := row.Scan(&totalCategoriesDB, &totalPriceDB); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to calculate total categories and prices")
	}

	// фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to commit transaction")
	}

	return totalItems, totalCategoriesDB, totalPriceDB, nil
}

// обрабатываем загрузку ZIP-архива с файлом data.csv
func loadArchive(ctx *gin.Context) {
	// получаем файл из запроса
	buf, err := getFileFromRequest(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// открываем ZIP-архив
	zipReader, err := openZipArchive(buf)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// извлекаем CSV-файл из архива
	csvFile, err := extractCSVFileFromZip(zipReader)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer csvFile.Close()

	// читаем данные из CSV
	records, err := readCSVFile(csvFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// обрабатываем данные и вставляем их в базу данных
	totalItems, totalCategoriesDB, totalPriceDB, err := processDatabaseTransaction(records)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// формируем ответ с итоговыми данными
	response := LoadResponse{
		TotalItems:      totalItems,
		TotalCategories: totalCategoriesDB,
		TotalPrice:      totalPriceDB,
	}

	// отправляем ответ
	ctx.JSON(http.StatusOK, response)
}

// строка таблицы цен
type PriceRow struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Price     float64   `json:"price"`
}

// функция для получения данных из базы данных
func fetchPricesFromDB() ([]PriceRow, error) {
	rows, err := db.Conn.Query("SELECT id, created_at, name, category, price FROM prices")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []PriceRow
	for rows.Next() {
		var item PriceRow
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.Name, &item.Category, &item.Price); err != nil {
			return nil, err
		}
		prices = append(prices, item)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return prices, nil
}

// функция для записи данных в CSV
func writeCSVToBuffer(prices []PriceRow) (*bytes.Buffer, error) {
	csvBuffer := &bytes.Buffer{}
	csvWriter := csv.NewWriter(csvBuffer)

	// записываем заголовок CSV
	err := csvWriter.Write([]string{"id", "create_date", "name", "category", "price"})
	if err != nil {
		return nil, err
	}

	// записываем данные в CSV
	for _, item := range prices {
		record := []string{
			fmt.Sprintf("%d", item.ID),
			item.CreatedAt.Format("2006-01-02"),
			item.Name,
			item.Category,
			fmt.Sprintf("%.2f", item.Price),
		}
		if err := csvWriter.Write(record); err != nil {
			return nil, err
		}
	}

	// завершаем запись CSV
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, err
	}

	return csvBuffer, nil
}

// функция для создания ZIP-архива
func createZipArchive(csvBuffer *bytes.Buffer) (*bytes.Buffer, error) {
	zipBuffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(zipBuffer)

	// добавляем CSV-файл в архив
	zipFile, err := zipWriter.Create("data.csv")
	if err != nil {
		return nil, err
	}

	_, err = zipFile.Write(csvBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	// завершаем создание архива
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return zipBuffer, nil
}

// обрабатываем выгрузку ZIP-архива
func uploadArchive(ctx *gin.Context) {
	// получаем данные из базы данных
	prices, err := fetchPricesFromDB()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query database"})
		return
	}

	// записываем данные в CSV
	csvBuffer, err := writeCSVToBuffer(prices)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv file"})
		return
	}

	// создаем ZIP-архив
	zipBuffer, err := createZipArchive(csvBuffer)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create zip file"})
		return
	}

	// отправляем архив в ответе
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", "attachment; filename=prices.zip")
	ctx.Data(http.StatusOK, "application/zip", zipBuffer.Bytes())
}
