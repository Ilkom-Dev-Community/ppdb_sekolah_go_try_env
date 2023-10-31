package controllers

import (
	"fmt"
	"ppdb_sekolah_go/configs"
	"ppdb_sekolah_go/constans"
	m "ppdb_sekolah_go/middlewares"
	"ppdb_sekolah_go/models"

	loger "log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"
)

func GetUsersController(c echo.Context) error {
	var users []models.User
	if err := configs.DB.Find(&users).Error; err != nil {
		log.Errorf("Failed to get users: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success get all users",
		constans.DATA:    users,
		//USAGE OF THE GLOBAL VARIABLE
	})
}

func GetUserController(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Invalid id: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid id")
	}
	var user models.User
	if err := configs.DB.First(&user, id).Error; err != nil {
		log.Errorf("Failed to get user with id %d: %s", id, err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success get user by id",
		constans.DATA:    user,
	})
}

func CreateUserController(c echo.Context) error {
	user := models.User{}
	if err := c.Bind(&user); err != nil {
		log.Errorf("Failed to bind request: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if IsEmailRegistered(user.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "Email address is already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}

	user.Password = string(hashedPassword)
	loger.Println(user)

	if err := configs.DB.Create(&user).Error; err != nil {
		log.Errorf("Failed to create user: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success create new user",
		constans.DATA:    user,
	})
}

// delete user by id
func DeleteUserController(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Errorf("Invalid id: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid id")
	}

	var user models.User
	if err := configs.DB.First(&user, id).Error; err != nil {
		log.Errorf("Failed to get user with id %d: %v", id, err)
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	if err := configs.DB.Delete(&user).Error; err != nil {
		log.Errorf("Failed to delete user with id %d: %v", id, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete user")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success deleted user",
	})
}

// update user by id
func UpdateUserController(c echo.Context) error {
	// get user id from url param
	userId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user id")
	}

	// get user by id
	var user models.User
	if err := configs.DB.First(&user, userId).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "User not found")
	}

	// bind request body to user struct
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// update password if new password is provided
	newPassword := c.FormValue("password")
	if newPassword != "" {
		// encrypt new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to encrypt password")
		}
		user.Password = string(hashedPassword)
	}

	// save user to database
	if err := configs.DB.Save(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success user updated",
		constans.DATA:    user,
	})
}

func LoginUserController(c echo.Context) error {
	user := models.User{}
	c.Bind(&user)

	err := configs.DB.Where("email = ?", user.Email).First(&user).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			constans.SUCCESS: false,
			constans.MESSAGE: "Failed to login",
			constans.ERROR:   err.Error(),
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(c.FormValue("password"))); err != nil {
		// fmt.Println(err)
		fmt.Println("pass :", c.FormValue("password"))
		fmt.Println("err :", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
	}

	fmt.Println("pass :", c.FormValue("password"))

	token, err := m.CreateToken(int(user.ID), user.Name, int(user.Role))
	fmt.Printf("UserID: %v, UserName: %v, UserRole: %v", user.ID, user.Name, user.Role)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			constans.SUCCESS: false,
			constans.MESSAGE: "Failed to login",
			constans.ERROR:   err.Error(),
		})
	}
	userResponse := models.UserResponse{user.ID, user.Name, user.Email, user.Role, token}

	return c.JSON(http.StatusOK, map[string]interface{}{
		constans.SUCCESS: true,
		constans.MESSAGE: "Success login",
		constans.DATA:    userResponse,
	})
}

func IsEmailRegistered(email string) bool {
	var user models.User
	if err := configs.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return false
	}
	return true
}
