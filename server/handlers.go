package server

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"project_sem/pkg/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type LoadResponse struct {
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}

func loadArchive(ctx *gin.Context) {
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
		return
	}

	defer file.Close()

	buf := new(bytes.Buffer)

	if _, err := io.Copy(buf, file); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid zip file"})
		return
	}

	var csvFile io.ReadCloser
	for _, file := range zipReader.File {
		if file.Name == "data.csv" {
			csvFile, err = file.Open()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open CSV file"})
				return
			}
			defer csvFile.Close()
			break
		}
	}

	if csvFile == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "data.csv not found in zip file"})
		return
	}

	reader := csv.NewReader(csvFile)
	if _, err := reader.Read(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to skip header in CSV file"})
		return
	}

	tx, err := db.Conn.Begin()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to begin transaction"})
		return
	}

	categorySet := make(map[string]struct{})
	var totalItems int
	var totalPrice float64

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read CSV file"})
			return
		}

		if len(record) < 5 {
			tx.Rollback()
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid record in CSV file"})
			return
		}

		name, category, price, createdAt := record[1], record[2], record[3], record[4]
		priceParse, err := strconv.ParseFloat(price, 64)
		if err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid price in CSV file - %s", price)})
			return
		}

		categorySet[category] = struct{}{}
		totalPrice += priceParse
		totalItems++

		if _, err := tx.Exec(
			"INSERT INTO prices (created_at, name, category, price) VALUES ($1, $2, $3, $4)",
			createdAt, name, category, price,
		); err != nil {
			tx.Rollback()
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data into the database"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	response := LoadResponse{
		TotalItems:      totalItems,
		TotalCategories: len(categorySet),
		TotalPrice:      totalPrice,
	}

	ctx.JSON(http.StatusOK, response)
}

type PriceRow struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Price     float64   `json:"price"`
}

func uploadArchive(ctx *gin.Context) {
	rows, err := db.Conn.Query("SELECT id, created_at, name, category, price FROM prices")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query database"})
		return
	}
	defer rows.Close()

	var prices []PriceRow
	for rows.Next() {
		var item PriceRow
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.Name, &item.Category, &item.Price); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan row"})
			return
		}
		prices = append(prices, item)
	}

	csvBuffer := &bytes.Buffer{}
	csvWriter := csv.NewWriter(csvBuffer)

	err = csvWriter.Write([]string{"id", "create_date", "name", "category", "price"})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv header"})
		return
	}

	for _, item := range prices {
		record := []string{
			fmt.Sprintf("%d", item.ID),
			item.CreatedAt.Format("2006-01-02"),
			item.Name,
			item.Category,
			fmt.Sprintf("%.2f", item.Price),
		}
		if err := csvWriter.Write(record); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write csv row"})
			return
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to flush csv writer"})
		return
	}

	zipBuffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(zipBuffer)

	zipFile, err := zipWriter.Create("data.csv")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create zip file"})
		return
	}

	_, err = zipFile.Write(csvBuffer.Bytes())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write to zip file"})
		return
	}

	if err := zipWriter.Close(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close zip writer"})
		return
	}

	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", "attachment; filename=prices.zip")
	ctx.Data(http.StatusOK, "application/zip", zipBuffer.Bytes())
}
