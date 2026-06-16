package transactiontypes

import (
	"math"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/proto/transaction_types_proto"
	"github.com/google/uuid"
)

func CreateTransactionType(req *transaction_types_proto.CreateTransactionTypeRequest) error {

	transactionType := TransactionType{
		ID:        uuid.New(),
		Name:      req.Name,
		Code:      req.Code,
		MaxAmount: req.MaxAmount,
		MinAmount: req.MinAmount,
		Status:    req.Status,
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Create(&transactionType).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetTransactionType(id string) (*TransactionType, error) {
	transactionType := TransactionType{}
	if err := global.GetDB().Model(&TransactionType{}).Where("id = ?", id).First(&transactionType).Error; err != nil {
		return nil, err
	}
	return &transactionType, nil
}

func GetTransactionTypes(req *transaction_types_proto.GetTransactionTypesRequest) (*transaction_types_proto.GetTransactionTypesResponse, error) {
	transactionTypes := []TransactionType{}
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	query := global.GetDB().Model(&TransactionType{})
	if req.SearchQuery != "" {
		query = query.Where("name LIKE ?", "%"+req.SearchQuery+"%")
	}
	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	query = query.Offset((page - 1) * pageSize).Limit(pageSize)

	if err := query.Find(&transactionTypes).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(pageSize)))

	var transactionTypesRes []*transaction_types_proto.TransactionType
	for _, transactionType := range transactionTypes {
		transactionTypesRes = append(transactionTypesRes, &transaction_types_proto.TransactionType{
			Id:        transactionType.ID.String(),
			Name:      transactionType.Name,
			Code:      transactionType.Code,
			MaxAmount: transactionType.MaxAmount,
			MinAmount: transactionType.MinAmount,
			Status:    transactionType.Status,
		})
	}
	return &transaction_types_proto.GetTransactionTypesResponse{
		TransactionType: transactionTypesRes,
		TotalPages:      totalPages,
		CurrentPage:     int32(page),
		HasMore:         int32(page) < totalPages,
	}, nil
}

func UpdateTransactionType(req *transaction_types_proto.EditTransactionTypeRequest) error {
	transactionType := TransactionType{}
	if err := global.GetDB().Model(&TransactionType{}).Where("id = ?", req.Id).First(&transactionType).Error; err != nil {
		return err
	}
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.MaxAmount != "" {
		updates["max_amount"] = req.MaxAmount
	}
	if req.MinAmount != "" {
		updates["min_amount"] = req.MinAmount
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&TransactionType{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func DeleteTransactionType(id string) error {
	transactionType := TransactionType{}
	if err := global.GetDB().Model(&TransactionType{}).Where("id = ?", id).First(&transactionType).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&SubTransactionType{}).Where("transaction_type_id = ?", id).Delete(&SubTransactionType{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&TransactionType{}).Where("id = ?", id).Delete(&transactionType).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func CreateSubTransactionType(req *transaction_types_proto.CreateSubTransactionTypeRequest) error {
	subTransactionType := SubTransactionType{
		ID:                uuid.New(),
		Name:              req.Name,
		Code:              req.Code,
		MaxAmount:         req.MaxAmount,
		MinAmount:         req.MinAmount,
		Status:            req.Status,
		TransactionTypeID: uuid.MustParse(req.TransactionTypeId),
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Create(&subTransactionType).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func UpdateSubTransactionType(req *transaction_types_proto.EditSubTransactionTypeRequest) error {
	subTransactionType := SubTransactionType{}
	if err := global.GetDB().Model(&SubTransactionType{}).Where("id = ?", req.Id).First(&subTransactionType).Error; err != nil {
		return err
	}
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.MaxAmount != "" {
		updates["max_amount"] = req.MaxAmount
	}
	if req.MinAmount != "" {
		updates["min_amount"] = req.MinAmount
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.TransactionTypeId != "" {
		updates["transaction_type_id"] = uuid.MustParse(req.TransactionTypeId)
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&SubTransactionType{}).Where("id = ?", req.Id).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func DeleteSubTransactionType(id string) error {
	subTransactionType := SubTransactionType{}
	if err := global.GetDB().Model(&SubTransactionType{}).Where("id = ?", id).First(&subTransactionType).Error; err != nil {
		return err
	}
	tx := global.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.Model(&SubTransactionType{}).Where("id = ?", id).Delete(&subTransactionType).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func GetSubTransactionTypes(req *transaction_types_proto.GetSubTransactionTypesRequest) (*transaction_types_proto.GetSubTransactionTypesResponse, error) {
	subTransactionTypes := []SubTransactionType{}
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	query := global.GetDB().Model(&SubTransactionType{})
	if req.SearchQuery != "" {
		query = query.Where("name LIKE ?", "%"+req.SearchQuery+"%")
	}
	if req.TransactionTypeId != "" {
		if _, err := uuid.Parse(req.TransactionTypeId); err == nil {
			query = query.Where("transaction_type_id = ?", req.TransactionTypeId)
		}
	}
	total := int64(0)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	query = query.Offset((page - 1) * pageSize).Limit(pageSize)

	if err := query.Find(&subTransactionTypes).Error; err != nil {
		return nil, err
	}
	totalPages := int32(math.Ceil(float64(total) / float64(pageSize)))

	var subTransactionTypesRes []*transaction_types_proto.SubTransactionType
	for _, subTransactionType := range subTransactionTypes {
		subTransactionTypesRes = append(subTransactionTypesRes, &transaction_types_proto.SubTransactionType{
			Id:                subTransactionType.ID.String(),
			Name:              subTransactionType.Name,
			Code:              subTransactionType.Code,
			MaxAmount:         subTransactionType.MaxAmount,
			MinAmount:         subTransactionType.MinAmount,
			Status:            subTransactionType.Status,
			TransactionTypeId: subTransactionType.TransactionTypeID.String(),
		})
	}
	return &transaction_types_proto.GetSubTransactionTypesResponse{
		SubTransactionType: subTransactionTypesRes,
		TotalPages:         totalPages,
		CurrentPage:        int32(page),
		HasMore:            int32(page) < totalPages,
	}, nil
}

func GetSubTransactionType(id string) (*SubTransactionType, error) {
	subTransactionType := SubTransactionType{}
	if err := global.GetDB().Model(&SubTransactionType{}).Where("id = ?", id).First(&subTransactionType).Error; err != nil {
		return nil, err
	}
	return &subTransactionType, nil
}
