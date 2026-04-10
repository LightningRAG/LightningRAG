package oauthapp

import (
	"errors"
	"strings"

	"github.com/LightningRAG/LightningRAG/server/global"
	"github.com/LightningRAG/LightningRAG/server/model/common/request"
	"github.com/LightningRAG/LightningRAG/server/model/system"
	systemReq "github.com/LightningRAG/LightningRAG/server/model/system/request"
	systemRes "github.com/LightningRAG/LightningRAG/server/model/system/response"
	idpoauth "github.com/LightningRAG/LightningRAG/server/oauth"
	"github.com/LightningRAG/LightningRAG/server/utils"
	"gorm.io/gorm"
)

type SysOAuthProviderService struct{}

var SysOAuthProviderServiceApp = new(SysOAuthProviderService)

func (s *SysOAuthProviderService) toAdmin(m system.SysOAuthProvider) systemRes.SysOAuthProviderAdmin {
	ex := map[string]interface{}(nil)
	if m.Extra != nil {
		ex = m.Extra
	}
	dn := strings.TrimSpace(m.DisplayName)
	if dn == "" {
		if p, err := idpoauth.Lookup(m.Kind); err == nil {
			dn = p.DefaultDisplayName()
		}
	}
	rawIcon := strings.TrimSpace(m.ButtonIcon)
	preview := rawIcon
	if preview == "" {
		preview = idpoauth.DefaultButtonIconForKind(m.Kind)
	}
	return systemRes.SysOAuthProviderAdmin{
		ID:                 m.ID,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		Kind:               m.Kind,
		Enabled:            m.Enabled,
		DisplayName:        dn,
		ButtonIcon:         rawIcon,
		ButtonIconPreview:  preview,
		ClientID:           m.ClientID,
		ClientSecretSet:    m.ClientSecretEnc != "",
		Scopes:             m.Scopes,
		Extra:              ex,
		DefaultAuthorityID: m.DefaultAuthorityID,
	}
}

func (s *SysOAuthProviderService) Create(req systemReq.SysOAuthProviderCreate) error {
	if global.LRAG_DB == nil {
		return errors.New("数据库未初始化")
	}
	kindNorm := strings.ToLower(strings.TrimSpace(req.Kind))
	if kindNorm == "" {
		return errors.New("kind 不能为空")
	}
	if _, err := idpoauth.Lookup(kindNorm); err != nil {
		return err
	}
	var existing system.SysOAuthProvider
	if err := global.LRAG_DB.Where("LOWER(kind) = ?", kindNorm).First(&existing).Error; err == nil {
		return errors.New("该登录方式已存在配置，请使用更新")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	enc, err := utils.EncryptOAuthSecret(req.ClientSecret)
	if err != nil {
		return err
	}
	bi := strings.TrimSpace(req.ButtonIcon)
	if err := idpoauth.ValidateButtonIcon(bi); err != nil {
		return err
	}
	authID := req.DefaultAuthorityID
	if authID == 0 {
		authID = 888
	}
	row := system.SysOAuthProvider{
		Kind:               kindNorm,
		Enabled:            req.Enabled,
		DisplayName:        strings.TrimSpace(req.DisplayName),
		ButtonIcon:         bi,
		ClientID:           strings.TrimSpace(req.ClientID),
		ClientSecretEnc:    enc,
		Scopes:             strings.TrimSpace(req.Scopes),
		Extra:              req.Extra,
		DefaultAuthorityID: authID,
	}
	return global.LRAG_DB.Create(&row).Error
}

func (s *SysOAuthProviderService) Update(req systemReq.SysOAuthProviderUpdate) error {
	if global.LRAG_DB == nil {
		return errors.New("数据库未初始化")
	}
	var row system.SysOAuthProvider
	if err := global.LRAG_DB.Where("id = ?", req.ID).First(&row).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.DisplayName != "" {
		updates["display_name"] = strings.TrimSpace(req.DisplayName)
	}
	if req.ClientID != "" {
		updates["client_id"] = strings.TrimSpace(req.ClientID)
	}
	if req.ClientSecret != "" {
		enc, err := utils.EncryptOAuthSecret(req.ClientSecret)
		if err != nil {
			return err
		}
		updates["client_secret_enc"] = enc
	}
	if req.Scopes != "" {
		updates["scopes"] = strings.TrimSpace(req.Scopes)
	}
	if req.Extra != nil {
		updates["extra"] = req.Extra
	}
	if req.DefaultAuthorityID != nil {
		updates["default_authority_id"] = *req.DefaultAuthorityID
	}
	if req.ButtonIcon != nil {
		v := strings.TrimSpace(*req.ButtonIcon)
		if err := idpoauth.ValidateButtonIcon(v); err != nil {
			return err
		}
		updates["button_icon"] = v
	}
	if len(updates) == 0 {
		return nil
	}
	return global.LRAG_DB.Model(&row).Updates(updates).Error
}

func (s *SysOAuthProviderService) Delete(id uint) error {
	if global.LRAG_DB == nil {
		return errors.New("数据库未初始化")
	}
	return global.LRAG_DB.Delete(&system.SysOAuthProvider{}, "id = ?", id).Error
}

func (s *SysOAuthProviderService) DeleteByIds(ids request.IdsReq) error {
	if global.LRAG_DB == nil {
		return errors.New("数据库未初始化")
	}
	return global.LRAG_DB.Delete(&[]system.SysOAuthProvider{}, "id in ?", ids.Ids).Error
}

func (s *SysOAuthProviderService) Find(id uint) (systemRes.SysOAuthProviderAdmin, error) {
	if global.LRAG_DB == nil {
		return systemRes.SysOAuthProviderAdmin{}, errors.New("数据库未初始化")
	}
	var row system.SysOAuthProvider
	if err := global.LRAG_DB.Where("id = ?", id).First(&row).Error; err != nil {
		return systemRes.SysOAuthProviderAdmin{}, err
	}
	return s.toAdmin(row), nil
}

func (s *SysOAuthProviderService) List(info systemReq.SysOAuthProviderSearch) (list []systemRes.SysOAuthProviderAdmin, total int64, err error) {
	if global.LRAG_DB == nil {
		return nil, 0, errors.New("数据库未初始化")
	}
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.LRAG_DB.Model(&system.SysOAuthProvider{})
	if info.Kind != "" {
		db = db.Where("kind LIKE ?", "%"+info.Kind+"%")
	}
	if info.Enabled != nil {
		db = db.Where("enabled = ?", *info.Enabled)
	}
	if err = db.Count(&total).Error; err != nil {
		return
	}
	var rows []system.SysOAuthProvider
	if limit > 0 {
		db = db.Limit(limit).Offset(offset)
	}
	err = db.Order("id desc").Find(&rows).Error
	if err != nil {
		return
	}
	for _, r := range rows {
		list = append(list, s.toAdmin(r))
	}
	return
}

// GetByKindForFlow 登录流程加载（含解密后的密钥）；kind 大小写不敏感
func (s *SysOAuthProviderService) GetByKindForFlow(kind string) (system.SysOAuthProvider, string, error) {
	var row system.SysOAuthProvider
	norm := strings.ToLower(strings.TrimSpace(kind))
	if norm == "" {
		return row, "", gorm.ErrRecordNotFound
	}
	if global.LRAG_DB == nil {
		return row, "", errors.New("数据库未初始化")
	}
	if err := global.LRAG_DB.Where("LOWER(kind) = ? AND enabled = ?", norm, true).First(&row).Error; err != nil {
		return row, "", err
	}
	sec, err := utils.DecryptOAuthSecret(row.ClientSecretEnc)
	if err != nil {
		return row, "", err
	}
	return row, sec, nil
}

// ListPublicEnabled 登录页：已启用且库中有配置
func (s *SysOAuthProviderService) ListPublicEnabled() ([]systemRes.OAuthPublicProvider, error) {
	if global.LRAG_DB == nil {
		return []systemRes.OAuthPublicProvider{}, nil
	}
	var rows []system.SysOAuthProvider
	if err := global.LRAG_DB.Where("enabled = ?", true).Find(&rows).Error; err != nil {
		return nil, err
	}
	var out []systemRes.OAuthPublicProvider
	for _, r := range rows {
		if _, err := idpoauth.Lookup(r.Kind); err != nil {
			continue
		}
		if strings.TrimSpace(r.ClientID) == "" || r.ClientSecretEnc == "" {
			continue
		}
		dn := strings.TrimSpace(r.DisplayName)
		if dn == "" {
			if p, err := idpoauth.Lookup(r.Kind); err == nil {
				dn = p.DefaultDisplayName()
			}
		}
		icon := strings.TrimSpace(r.ButtonIcon)
		if icon == "" {
			icon = idpoauth.DefaultButtonIconForKind(r.Kind)
		}
		out = append(out, systemRes.OAuthPublicProvider{
			Kind:        strings.ToLower(strings.TrimSpace(r.Kind)),
			DisplayName: dn,
			ButtonIcon:  icon,
		})
	}
	return out, nil
}

// RegisteredKindsForAdmin 已注册实现 + 默认展示名
func (s *SysOAuthProviderService) RegisteredKindsForAdmin() []systemRes.OAuthRegisteredKind {
	kinds := idpoauth.RegisteredKinds()
	out := make([]systemRes.OAuthRegisteredKind, 0, len(kinds))
	for _, k := range kinds {
		p, err := idpoauth.Lookup(k)
		if err != nil {
			continue
		}
		out = append(out, systemRes.OAuthRegisteredKind{
			Kind:              k,
			DisplayName:       p.DefaultDisplayName(),
			DefaultButtonIcon: idpoauth.DefaultButtonIconForKind(k),
		})
	}
	return out
}
