package handlers

import "github.com/gofiber/fiber/v2"

func GetIP(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"ip": c.IP()})
}

func GetHeaders(c *fiber.Ctx) error {
	return c.JSON(c.GetReqHeaders())
}

// diag.Post("/mail", func(c *fiber.Ctx) error {
// 	type EmailInput struct {
// 		Subject string `json:"subject" validate:"required"`
// 		Body    string `json:"body" validate:"required"`
// 		From    string `json:"from" validate:"required"`
// 		To      string `json:"to" validate:"required"`
// 	}

// 	var input EmailInput
// 	if err := c.BodyParser(&input); err != nil {
// 		return c.SendStatus(fiber.StatusBadRequest)
// 	}

// 	err := config.Validate.Struct(input)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
// 	}

// 	message := mail.Email{
// 		Subject: input.Subject,
// 		Body:    input.Body,
// 		From:    input.From,
// 		To:      []string{input.To},
// 	}

// 	mailer := mail.NewMailer(cfg.MailgunDomain, cfg.MailgunAPIKey, cfg.MailgunAPIBase)
// 	mailer.SendMail(&message)

// 	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Message queued"})
// })
