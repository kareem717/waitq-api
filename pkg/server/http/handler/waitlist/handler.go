package waitlist

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"time"

	"waitq/api/pkg/entities/waitlist"
	helper "waitq/api/pkg/server/http/handler/shared"
	"waitq/api/pkg/service"
	"waitq/api/pkg/storage/postgres/shared"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type httpHandler struct {
	waitlistService service.WaitlistService
	logger          *zap.Logger
}

func newHTTPHandler(waitlistService service.WaitlistService, logger *zap.Logger) *httpHandler {
	return &httpHandler{
		waitlistService: waitlistService,
		logger:          logger,
	}
}

type GetWaitlistByIDInput struct {
	ID string `path:"id"`
}

type GetWaitlistByIDOutput struct {
	Body struct {
		Message  string             `json:"message"`
		Waitlist *waitlist.Waitlist `json:"waitlist"`
	}
}

func (h *httpHandler) getByID(ctx context.Context, input *GetWaitlistByIDInput) (*GetWaitlistByIDOutput, error) {
	accountID, err := uuid.Parse(input.ID) // fetching and validation input
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid waitlist ID")
	}

	waitlist, err := h.waitlistService.GetById(ctx, accountID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlist.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlist))
		return nil, huma.Error403Forbidden("Cannot access waitlist")
	}

	resp := &GetWaitlistByIDOutput{}
	resp.Body.Message = "Waitlist fetched successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type CreateWaitlistInput struct {
	Body struct {
		CreateWaitlistFields struct {
			Name      string    `json:"name" minLength:"3" maxLength:"100"`
			AccountID uuid.UUID `json:"accountId" minLength:"36" maxLength:"36" format:"uuid"`
		} `json:"waitlist"`
	}
}

type CreateWaitlistOutput struct {
	Body struct {
		Message  string             `json:"message"`
		Waitlist *waitlist.Waitlist `json:"waitlist"`
	}
}

func (h *httpHandler) create(ctx context.Context, input *CreateWaitlistInput) (*CreateWaitlistOutput, error) {
	if account := helper.GetAuthenticatedAccount(ctx); account.ID != input.Body.CreateWaitlistFields.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", account), zap.Any("input", input))
		return nil, huma.Error403Forbidden("Cannot create waitlist for another account")
	}

	waitlist, err := h.waitlistService.Create(ctx, waitlist.Waitlist{
		Name:      input.Body.CreateWaitlistFields.Name,
		AccountID: input.Body.CreateWaitlistFields.AccountID,
	})

	if err != nil {
		h.logger.Error("failed to create waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while creating the waitlist")
	}

	resp := &CreateWaitlistOutput{}
	resp.Body.Message = "Waitlist created successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type GetWaitlistByAccountIDInput struct {
	AccountID string `path:"accountId"`
	Body      struct {
		PaginationParams shared.PaginationRequest `json:"paginationParams"`
	}
}

type GetWaitlistByAccountIDOutput struct {
	Body struct {
		Count    int                  `json:"count"`
		Message  string               `json:"message"`
		Waitlist *[]waitlist.Waitlist `json:"waitlists"`
	}
}

func (h *httpHandler) getByAccountID(ctx context.Context, input *GetWaitlistByAccountIDInput) (*GetWaitlistByAccountIDOutput, error) {
	accountID, err := uuid.Parse(input.AccountID) // fetching and validation input
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid account ID")
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != accountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("accountID", accountID))
		return nil, huma.Error403Forbidden("Cannot access accounts")
	}

	waitlists, err := h.waitlistService.GetByAccountId(ctx, accountID, input.Body.PaginationParams)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Accounts not found")
		default:
			h.logger.Error("failed to fetch accounts", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the accounts")
		}
	}

	resp := &GetWaitlistByAccountIDOutput{}
	resp.Body.Message = "Waitlists fetched successfully"
	resp.Body.Waitlist = &waitlists
	resp.Body.Count = len(*resp.Body.Waitlist)

	return resp, nil
}

type UpdateWaitlistInput struct {
	ID   uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Body struct {
		UpdateWaitlistFields struct {
			Name string `json:"name" minLength:"3" maxLength:"100"`
		} `json:"waitlist"`
	}
}

type UpdateWaitlistOutput struct {
	Body struct {
		Message  string             `json:"message"`
		Waitlist *waitlist.Waitlist `json:"waitlist"`
	}
}

func (h *httpHandler) update(ctx context.Context, input *UpdateWaitlistInput) (*UpdateWaitlistOutput, error) {
	waitlistRecord, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistRecord.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistRecord.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistRecord))
		return nil, huma.Error403Forbidden("Cannot update waitlist for another account")
	}

	waitlist, err := h.waitlistService.Update(ctx, input.ID, waitlist.Waitlist{
		Name: input.Body.UpdateWaitlistFields.Name,
	})

	if err != nil {
		h.logger.Error("failed to update waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while updating the waitlist")
	}

	resp := &UpdateWaitlistOutput{}
	resp.Body.Message = "Waitlist updated successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type UpdateWaitlistJWTSecretInput struct {
	ID   uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Body struct {
		UpdateWaitlistJWTSecretFields struct {
			JWTSecret string `json:"jwtSecret" minLength:"32" maxLength:"512" required:"false"`
		} `json:"waitlist"`
	}
}

type UpdateWaitlistJWTSecretOutput struct {
	Body struct {
		Message  string             `json:"message"`
		Waitlist *waitlist.Waitlist `json:"waitlist"`
	}
}

func (h *httpHandler) updateJWTSecret(ctx context.Context, input *UpdateWaitlistJWTSecretInput) (*UpdateWaitlistJWTSecretOutput, error) {
	waitlistRecord, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if waitlistRecord.DeletedAt != nil {
		return nil, huma.Error404NotFound("Waitlist not found")
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistRecord.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistRecord))
		return nil, huma.Error403Forbidden("Cannot update waitlist for another account")
	}

	waitlist, err := h.waitlistService.UpdateJWTSecret(ctx, input.ID, input.Body.UpdateWaitlistJWTSecretFields.JWTSecret)
	if err != nil {
		h.logger.Error("failed to update waitlist jwt secret", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while updating the waitlist jwt secret")
	}

	resp := &UpdateWaitlistJWTSecretOutput{}
	resp.Body.Message = "Waitlist JWT secret updated successfully"
	resp.Body.Waitlist = &waitlist

	return resp, nil
}

type DeleteWaitlistInput struct {
	ID uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
}

type DeleteWaitlistOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func (h *httpHandler) delete(ctx context.Context, input *DeleteWaitlistInput) (*DeleteWaitlistOutput, error) {
	waitlist, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Account not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlist.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlist))
		return nil, huma.Error403Forbidden("Cannot delete waitlist for another account")
	}

	err = h.waitlistService.Delete(ctx, input.ID)

	if err != nil {
		h.logger.Error("failed to delete account", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while deleting the account")
	}

	resp := &DeleteWaitlistOutput{}
	resp.Body.Message = "Waitlist deleted successfully"

	return resp, nil
}

type AddEmailsInput struct {
	ID   uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Body struct {
		Emails []string `json:"emails" minLength:"3" maxLength:"320" format:"email"`
	}
}

type AddEmailsOutput struct {
	Body struct {
		Message     string           `json:"message"`
		EmailsAdded []waitlist.Email `json:"emailsAdded"`
		Count       int              `json:"count"`
	}
}

func (h *httpHandler) addEmails(ctx context.Context, input *AddEmailsInput) (*AddEmailsOutput, error) {

	if len(input.Body.Emails) == 0 {
		return nil, huma.Error400BadRequest("No emails provided")
	}

	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	//TODO: this doesn't work beacuase of no middleware
	// Anon users can only add one email at a time
	if len(input.Body.Emails) > 1 {
		if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
			h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
			return nil, huma.Error403Forbidden("Bearer token is required to add multiple emails")
		}
	}

	newEmails := make([]waitlist.Email, len(input.Body.Emails))

	for i, emailInput := range input.Body.Emails {
		newEmails[i] = waitlist.Email{
			WaitlistID: waitlistResp.ID,
			Email:      emailInput,
		}
	}

	emails, err := h.waitlistService.AddEmails(ctx, newEmails)
	if err != nil {
		h.logger.Error("failed to add emails to waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while adding emails to the waitlist")
	}

	resp := &AddEmailsOutput{}
	resp.Body.Message = "Emails added to waitlist successfully"
	resp.Body.EmailsAdded = emails
	resp.Body.Count = len(emails)

	return resp, nil
}

type DeleteEmailInput struct {
	ID    uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Email string    `query:"email" minLength:"3" maxLength:"320" format:"email"`
}

type DeleteEmailsOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func (h *httpHandler) deleteEmail(ctx context.Context, input *DeleteEmailInput) (*DeleteEmailsOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
		return nil, huma.Error403Forbidden("Cannot delete email from waitlist for another account")
	}

	err = h.waitlistService.DeleteEmail(ctx, input.ID, input.Email)
	if err != nil {
		h.logger.Error("failed to delete email from waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while deleting email from the waitlist")
	}

	resp := &DeleteEmailsOutput{}
	resp.Body.Message = "Email deleted from waitlist successfully"
	return resp, nil
}

type UpdateEmailInput struct {
	ID   uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Body struct {
		Email waitlist.Email `json:"email"`
	}
}

type UpdateEmailOutput struct {
	Body struct {
		Message      string         `json:"message"`
		EmailUpdated waitlist.Email `json:"emailUpdated"`
	}
}

func (h *httpHandler) updateEmail(ctx context.Context, input *UpdateEmailInput) (*UpdateEmailOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
		return nil, huma.Error403Forbidden("Cannot update email in waitlist for another account")
	}

	if input.Body.Email.WaitlistID != input.ID {
		return nil, huma.Error400BadRequest("Email ID does not match waitlist ID")
	}

	email, err := h.waitlistService.UpdateEmail(ctx, input.Body.Email)
	if err != nil {
		h.logger.Error("failed to update email", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while updating the email")
	}

	resp := &UpdateEmailOutput{}
	resp.Body.Message = "Email updated successfully"
	resp.Body.EmailUpdated = email

	return resp, nil
}

type GetEmailsByWaitlistIDInput struct {
	ID   uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Body struct {
		PaginationParams shared.EmailPaginationRequest `json:"paginationParams"`
	}
}

type GetEmailsByWaitlistIDOutput struct {
	Body struct {
		Message string           `json:"message"`
		Emails  []waitlist.Email `json:"emails"`
		Count   int              `json:"count"`
	}
}

func (h *httpHandler) getEmailsByWaitlistID(ctx context.Context, input *GetEmailsByWaitlistIDInput) (*GetEmailsByWaitlistIDOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
		return nil, huma.Error403Forbidden("Cannot add emails to waitlist for another account")
	}

	emails, err := h.waitlistService.GetEmailsByWaitlistID(ctx, input.ID, input.Body.PaginationParams)
	if err != nil {
		h.logger.Error("failed to fetch emails", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while fetching the emails")
	}

	resp := &GetEmailsByWaitlistIDOutput{}
	resp.Body.Message = "Emails added to waitlist successfully"
	resp.Body.Emails = emails
	resp.Body.Count = len(emails)

	return resp, nil
}

type UnsubscribeEmailInput struct {
	ID   uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Body struct {
		EncodedEmail string `json:"encodedEmail"`
	}
}

type UnsubscribeEmailOutput struct {
	Body struct {
		Message string         `json:"message"`
		Email   waitlist.Email `json:"email"`
	}
}

func (h *httpHandler) unsubscribeEmail(ctx context.Context, input *UnsubscribeEmailInput) (*UnsubscribeEmailOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	token, err := jwt.Parse(input.Body.EncodedEmail, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			h.logger.Error("unexpected signing method", zap.Any("token", token.Header["alg"]))
			return nil, huma.Error400BadRequest("Invalid encoded email")
		}

		return []byte(waitlistResp.JWTSecret), nil
	})
	if err != nil {
		h.logger.Error("failed to parse token", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while parsing the encoded email")
	}

	var inputEmail, waitlistID string
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		inputEmail = claims["email"].(string)
		waitlistID = claims["ref"].(string)
	} else {
		h.logger.Error("failed to parse token", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while parsing the encoded email")
	}

	if waitlistID != input.ID.String() {
		return nil, huma.Error400BadRequest("Invalid waitlist ID")
	}

	email, err := h.waitlistService.GetEmailsByWaitlistIDAndEmail(ctx, waitlistResp.ID, inputEmail)
	if err != nil {
		h.logger.Error("failed to fetch email", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while fetching the email")
	}

	if email.WaitlistID != waitlistResp.ID {
		return nil, huma.Error400BadRequest("Email does not belong to waitlist")
	}

	if email.DeletedAt != nil {
		return nil, huma.Error400BadRequest("Email has already been removed from waitlist")
	}

	if email.UnsubscribedAt != nil {
		return nil, huma.Error400BadRequest("Email has already been unsubscribed")
	}

	emailResp, err := h.waitlistService.UnsubscribeEmail(ctx, waitlistResp.ID, inputEmail)
	if err != nil {
		h.logger.Error("failed to unsubscribe email from waitlist", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while unsubscribing email from the waitlist")
	}

	resp := &UnsubscribeEmailOutput{}
	resp.Body.Message = "Email unsubscribed from waitlist successfully"
	resp.Body.Email = emailResp

	return resp, nil
}

type GetUnsubscribedEmailJWTInput struct {
	ID    uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
	Email string    `query:"email" minLength:"3" maxLength:"320" format:"email"`
}

type GetUnsubscribedEmailJWTOutput struct {
	Body struct {
		Message string `json:"message"`
		Encoded string `json:"encoded"`
	}
}

func (h *httpHandler) getUnsubscribedEmailJWT(ctx context.Context, input *GetUnsubscribedEmailJWTInput) (*GetUnsubscribedEmailJWTOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
		return nil, huma.Error403Forbidden("Cannot get unsubscribed email JWT for another account")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   "waitq",
		"ref":   input.ID.String(),
		"email": input.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 365 * 10).Unix(),
		"iat":   time.Now().Unix(),
	})

	// Ensure the secret is a byte slice for HMAC
	hmacSecret := []byte(waitlistResp.JWTSecret)

	// Sign and get the complete encoded token as a string using the HMAC secret
	encoded, err := token.SignedString(hmacSecret)
	if err != nil {
		h.logger.Error("failed to sign token", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while signing the token")
	}

	resp := &GetUnsubscribedEmailJWTOutput{}
	resp.Body.Message = "Unsubscribed email JWT created successfully"
	resp.Body.Encoded = encoded

	return resp, nil
}

type GetWaitlistAnalyticsInput struct {
	ID uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
}

type GetWaitlistAnalyticsOutput struct {
	Body struct {
		Message   string                     `json:"message"`
		Analytics waitlist.WaitlistAnalytics `json:"analytics"`
	}
}

func (h *httpHandler) getWaitlistAnalytics(ctx context.Context, input *GetWaitlistAnalyticsInput) (*GetWaitlistAnalyticsOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
		return nil, huma.Error403Forbidden("Cannot get analytics for another account")
	}

	analytics, err := h.waitlistService.GetAnalytics(ctx, waitlistResp.ID)
	if err != nil {
		h.logger.Error("failed to fetch analytics", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while fetching the email")
	}

	resp := &GetWaitlistAnalyticsOutput{}
	resp.Body.Message = "Waitlist analytics fetched successfully"
	resp.Body.Analytics = analytics

	return resp, nil

}

type ExportEmailsToCSVInput struct {
	ID uuid.UUID `path:"id" minLength:"36" maxLength:"36" format:"uuid"`
}

type ExportEmailsToCSVOutput struct {
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	Body               []byte
}

func (h *httpHandler) exportEmailsToCSV(ctx context.Context, input *ExportEmailsToCSVInput) (*ExportEmailsToCSVOutput, error) {
	waitlistResp, err := h.waitlistService.GetById(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, huma.Error404NotFound("Waitlist not found")
		default:
			h.logger.Error("failed to fetch waitlist", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while fetching the waitlist")
		}
	}

	if ctxAccount := helper.GetAuthenticatedAccount(ctx); ctxAccount.ID != waitlistResp.AccountID {
		h.logger.Error("unauthorized access", zap.Any("account", ctxAccount), zap.Any("waitlist", waitlistResp))
		return nil, huma.Error403Forbidden("Cannot export emails to CSV for another account")
	}

	emailsChan, errChan := h.waitlistService.ExportEmails(ctx, input.ID)

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	for email := range emailsChan {
		if err := writer.Write([]string{email}); err != nil {
			h.logger.Error("failed to write email to CSV", zap.Error(err))
			return nil, huma.Error500InternalServerError("An error occurred while writing emails to CSV")
		}
		writer.Flush()
	}

	if err := <-errChan; err != nil {
		h.logger.Error("failed to export emails", zap.Error(err))
		return nil, huma.Error500InternalServerError("An error occurred while exporting emails")
	}

	resp := &ExportEmailsToCSVOutput{}
	resp.ContentType = "text/csv"
	resp.ContentDisposition = fmt.Sprintf("attachment; filename=%s.csv", waitlistResp.Name)
	resp.Body = buffer.Bytes()

	return resp, nil
}
