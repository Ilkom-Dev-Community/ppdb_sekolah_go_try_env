package controllers

import (
	"context"
	"io"
	loger "log"
	"net/http"
	"ppdb_sekolah_go/configs"
	"ppdb_sekolah_go/constans"
	"ppdb_sekolah_go/models"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func GetDatapokokController(c echo.Context) error {
	var users []models.Datapokok
	if err := configs.DB.Find(&users).Error; err != nil {
		log.Errorf("Failed to get datapokok: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success get all datapokok",
		constans.DATA:    users,
	})
}

func GetDatapokokControllerByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Invalid id: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid id")
	}

	var user models.Datapokok
	if err := configs.DB.First(&user, id).Error; err != nil {
		log.Errorf("Failed to get datapokok with id %d: %s", id, err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var nilai models.Nilai
	if err := configs.DB.Where("datapokok_id = ?", id).First(&nilai).Error; err != nil {
		log.Errorf("Failed to get nilai with datapokok_id %d: %s", id, err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user.Nilai = append(user.Nilai, nilai)

	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success get datapokok by ID",
		constans.DATA:    user,
	})
}

func CreateDatapokokController(c echo.Context, client *storage.Client, bucketName string) error {
	// Create a request structure that includes Datapokok and Nilai data
	requestData := struct {
		Datapokok models.Datapokok `json:"datapokok"`
		Nilai     models.Nilai     `json:"nilai"`
	}{}

	// Bind the request data from the JSON body
	if err := c.Bind(&requestData); err != nil {
		log.Errorf("Failed to bind request: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())

	}

	userIDDatapokokStr := c.FormValue("user_id")
	userIDDatapokok, err := strconv.ParseUint(userIDDatapokokStr, 10, 0)
	if err != nil {
		log.Errorf("Failed to convert user_id to a uint: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id")
	}

	requestData.Datapokok.UserID = uint64(userIDDatapokok)

	requestData.Datapokok.Email = c.FormValue("email")
	requestData.Datapokok.NamaLengkap = c.FormValue("nama_lengkap")
	requestData.Datapokok.NISN = c.FormValue("nisn")
	requestData.Datapokok.JenisKelamin = c.FormValue("jenis_kelamin")
	requestData.Datapokok.TempatLahir = c.FormValue("tempat_lahir")

	// Date of birth handling
	dobStr := c.FormValue("tanggal_lahir")
	dob, err := time.Parse("2006-01-02", dobStr)
	if err == nil {
		requestData.Datapokok.TanggalLahir = &dob
	}

	requestData.Datapokok.AsalSekolah = c.FormValue("asal_sekolah")
	requestData.Datapokok.NamaAyah = c.FormValue("nama_ayah")
	requestData.Datapokok.NoWaAyah = c.FormValue("no_wa_ayah")
	requestData.Datapokok.NamaIbu = c.FormValue("nama_ibu")
	requestData.Datapokok.NoWaIbu = c.FormValue("no_wa_ibu")

	// Create the Datapokok record in the database
	if err := configs.DB.Create(&requestData.Datapokok).Error; err != nil {
		log.Errorf("Failed to create datapokok: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Handle file upload
	image, err := c.FormFile("pas_foto")
	if err != nil {
		log.Errorf("Failed to get the image file: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Image upload failed")
	}

	// Generate a unique filename using a UUID
	uniqueFilename := uuid.NewString()

	// Upload the image to the existing Google Cloud Storage bucket
	ctx := context.Background()
	wc := client.Bucket(bucketName).Object(uniqueFilename).NewWriter(ctx)
	defer wc.Close()

	src, err := image.Open()
	if err != nil {
		log.Errorf("Failed to open the image file: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process image")
	}
	defer src.Close()

	if _, err = io.Copy(wc, src); err != nil {
		log.Errorf("Failed to copy the image to the bucket: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upload image")
	}

	requestData.Datapokok.PasFoto = "https://storage.googleapis.com/" + bucketName + "/" + uniqueFilename

	// Now requestData.Datapokok.ID contains the ID of the newly created Datapokok record
	loger.Println("Created Datapokok with ID:", requestData.Datapokok.ID)

	// Set the Nilai's DatapokokID to the ID of the created Datapokok record
	requestData.Nilai.DataPokokID = requestData.Datapokok.ID
	requestData.Nilai.BahasaIndonesia = 0
	requestData.Nilai.IlmuPengetahuanAlam = 0
	requestData.Nilai.Matematika = 0
	requestData.Nilai.TestMembacaAlQuran = 0
	requestData.Nilai.Status = "BELUM LULUS"

	// Create the Nilai record in the database
	if err := configs.DB.Create(&requestData.Nilai).Error; err != nil {
		log.Errorf("Failed to create nilai: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// requestData.Nilai.Utama

	requestData.Datapokok.Nilai = append(requestData.Datapokok.Nilai, requestData.Nilai)

	// Return a response
	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success create new Datapokok and Nilai",
		constans.DATA:    requestData.Datapokok,
	})
}

// delete user by id
func DeleteDatapokokController(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Invalid id: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid id")
	}

	var user models.Datapokok
	if err := configs.DB.First(&user, id).Error; err != nil {
		log.Errorf("Failed to get datapokok with id %d: %v", id, err)
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	if err := configs.DB.Delete(&user).Error; err != nil {
		log.Errorf("Failed to delete datapokok with id %d: %v", id, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete datapokok")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "success deleted datapokok",
	})
}

// update user by id
func UpdateDatapokokController(c echo.Context) error {
	// get user id from url param
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid datapokok id")
	}

	// get user by id
	var user models.Datapokok
	if err := configs.DB.First(&user, userId).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "datapokok not found")
	}

	// bind request body to user struct
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// save user to database
	if err := configs.DB.Save(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success datapokok updated",
		constans.DATA:    user,
	})
}
