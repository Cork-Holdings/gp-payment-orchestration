package paymentservices

import (
	"math"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/payment_services_proto"
	"github.com/google/uuid"
)

func CreatePaymentService(req *payment_services_proto.CreatePaymentServiceRequest) error {

	paymentService := PaymentService{
		ID:     uuid.New(),
		Name:   req.Name,
		Status: req.Status,
	}

	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&paymentService).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetPaymentService(id string) (*PaymentService, error) {
	paymentService := PaymentService{}
	if err := global.GetDB().Model(&PaymentService{}).Where("id = ?", id).First(&paymentService).Error; err != nil {
		return nil, err
	}
	return &paymentService, nil
}

func GetPaymentServices(req *payment_services_proto.GetPaymentServicesRequest) (*payment_services_proto.GetPaymentServicesResponse, error) {
	paymentServices := []PaymentService{}
	query := global.GetDB().Model(&PaymentService{})
	if req.SearchQuery != "" {
		query = query.Where("name LIKE ?", "%"+req.SearchQuery+"%")
	}

	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if req.Page != 0 {
		query = query.Offset(int((req.Page - 1) * req.PageSize))
	}
	if req.PageSize != 0 {
		query = query.Limit(int(req.PageSize))
	}
	if err := query.Find(&paymentServices).Error; err != nil {
		return nil, err
	}

	var paymentServiceRes []*payment_services_proto.PaymentService
	for _, paymentService := range paymentServices {
		paymentServiceRes = append(paymentServiceRes, &payment_services_proto.PaymentService{
			Id:     paymentService.ID.String(),
			Name:   paymentService.Name,
			Status: paymentService.Status,
			Logo:   paymentService.Logo,
		})
	}
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))
	return &payment_services_proto.GetPaymentServicesResponse{
		PaymentService: paymentServiceRes,
		TotalPages:     totalPages,
		CurrentPage:    req.Page,
		HasMore:        req.Page < totalPages,
	}, nil
}

func UpdatePaymentService(req *payment_services_proto.EditPaymentServiceRequest) error {
	paymentService := PaymentService{}
	if err := global.GetDB().Model(&PaymentService{}).Where("id = ?", req.Id).First(&paymentService).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if err := global.GetDB().Model(&PaymentService{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&PaymentService{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeletePaymentService(id string) error {
	paymentService := PaymentService{}
	if err := global.GetDB().Model(&PaymentService{}).Where("id = ?", id).First(&paymentService).Delete(&PaymentService{}).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&PaymentService{}).Where("id = ?", id).Delete(&PaymentService{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
