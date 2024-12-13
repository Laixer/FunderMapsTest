package main

import (
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"

	"fundermaps/internal/database"
	"fundermaps/pkg/utils"
)

const (
	apiBaseURL    = "https://goldfish-app-4m6mn.ondigitalocean.app/api"
	apiManagement = "v1/management"
)

var apiKey string

type ResponseError struct {
	Message string `json:"message"`
}

var apiServiceBase = func() *resty.Client {
	return resty.New().
		SetBaseURL(apiBaseURL).
		SetHeader("Accept", "application/json").
		SetHeader("X-API-Key", apiKey).
		SetError(&ResponseError{}).
		OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
			if resp.StatusCode() >= 400 {
				return fmt.Errorf(resp.Error().(*ResponseError).Message)
			}

			return nil
		})
}

var rootCmd = &cobra.Command{
	Use:   "fundermaps",
	Short: "Fundermaps CLI",
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

var userCreateCmd = &cobra.Command{
	Use:   "create <email>",
	Short: "Create a new user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]
		password := utils.GenerateRandomString(12)

		resp, err := apiServiceBase().R().
			SetBody(map[string]string{
				"email":    email,
				"password": password,
			}).
			SetResult(&database.User{}).
			Post(apiManagement + "/user")

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		user := resp.Result().(*database.User)

		fmt.Println("User ID  :", user.ID)
		fmt.Println("Email    :", user.Email)
		fmt.Println("Role     :", user.Role)
		fmt.Println("Password :", password)
	},
}

var userCreateAuthKeyCmd = &cobra.Command{
	Use:   "auth-key <user_id>",
	Short: "Create a new auth key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userID := args[0]

		resp, err := apiServiceBase().R().
			SetBody(map[string]string{
				"user_id": userID,
			}).
			SetResult(&database.AuthKey{}).
			Post(fmt.Sprintf("%s/user/%s/auth-token", apiManagement, userID))

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		user := resp.Result().(*database.AuthKey)

		fmt.Println("User ID :", user.UserID)
		fmt.Println("Key     :", user.Key)
	},
}

var userResetPasswordCmd = &cobra.Command{
	Use:   "reset-password <user_id>",
	Short: "Reset user password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		userID := args[0]
		password := utils.GenerateRandomString(12)

		resp, err := apiServiceBase().R().
			SetBody(map[string]string{
				"user_id":  userID,
				"password": password,
			}).
			SetResult(&database.User{}).
			Post(fmt.Sprintf("%s/user/%s/reset-password", apiManagement, userID))

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		user := resp.Result().(*database.User)

		fmt.Println("User ID :", user.ID)
		fmt.Println("Password:", password)
	},
}

// TODO: This is just a validation of the API key
var userProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get user profile",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := apiServiceBase().R().
			SetResult(&database.User{}).
			Get("/user/me")

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		user := resp.Result().(*database.User)

		fmt.Println("User ID :", user.ID)
		fmt.Println("Email   :", user.Email)
		fmt.Println("Role    :", user.Role)
		fmt.Println("\nOrganizations")
		for _, org := range user.Organizations {
			fmt.Println("  - Org ID :", org.ID)
			fmt.Println("    Name   :", org.Name)
		}
	},
}

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organizations",
}

var orgCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new organization",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		resp, err := apiServiceBase().R().
			SetBody(map[string]string{
				"name": name,
			}).
			SetResult(&database.Organization{}).
			Post(apiManagement + "/org")

		fmt.Println("Status Code:", resp.StatusCode())

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		org := resp.Result().(*database.Organization)

		fmt.Println("Org ID :", org.ID)
		fmt.Println("Name   :", org.Name)
	},
}

func main() {
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userCreateAuthKeyCmd)
	userCmd.AddCommand(userResetPasswordCmd)
	userCmd.AddCommand(userProfileCmd)
	orgCmd.AddCommand(orgCreateCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(orgCmd)

	rootCmd.PersistentFlags().StringVarP(&apiKey, "key", "k", "", "API key")
	rootCmd.MarkPersistentFlagRequired("key")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
