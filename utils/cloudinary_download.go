package utils

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
)

type CloudinaryAssetRef struct {
	ResourceType string
	DeliveryType string
	PublicID     string
	Format       string
}

// ParseCloudinaryDeliveryURL extracts Cloudinary asset metadata from a delivery URL.
func ParseCloudinaryDeliveryURL(rawURL string) (*CloudinaryAssetRef, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	host := strings.ToLower(u.Host)
	if !strings.Contains(host, "cloudinary.com") {
		return nil, fmt.Errorf("not a cloudinary url")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected cloudinary path")
	}

	resourceType := parts[1]
	deliveryType := parts[2]
	assetSegments := parts[3:]
	if len(assetSegments) == 0 {
		return nil, fmt.Errorf("missing asset path")
	}

	startIdx := 0
	for i, seg := range assetSegments {
		if strings.HasPrefix(seg, "v") && len(seg) > 1 {
			if _, convErr := strconv.Atoi(seg[1:]); convErr == nil {
				startIdx = i + 1
				break
			}
		}
	}
	if startIdx >= len(assetSegments) {
		return nil, fmt.Errorf("missing public id")
	}

	publicPath := strings.Join(assetSegments[startIdx:], "/")
	ext := strings.TrimPrefix(path.Ext(publicPath), ".")
	if ext == "" {
		return nil, fmt.Errorf("missing format extension")
	}

	publicID := strings.TrimSuffix(publicPath, "."+ext)
	if publicID == "" {
		return nil, fmt.Errorf("missing public id")
	}

	return &CloudinaryAssetRef{
		ResourceType: resourceType,
		DeliveryType: deliveryType,
		PublicID:     publicID,
		Format:       ext,
	}, nil
}

// BuildCloudinaryTemporaryDownloadURL creates a signed, time-limited download URL.
func BuildCloudinaryTemporaryDownloadURL(cld *cloudinary.Cloudinary, rawURL string, ttl time.Duration) (string, error) {
	if cld == nil {
		return "", fmt.Errorf("cloudinary client is not initialized")
	}

	assetRef, err := ParseCloudinaryDeliveryURL(rawURL)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	expiresAt := time.Now().Add(ttl).Unix()

	params := url.Values{}
	params.Set("public_id", assetRef.PublicID)
	params.Set("format", assetRef.Format)
	if assetRef.DeliveryType != "" {
		params.Set("type", assetRef.DeliveryType)
	}
	params.Set("timestamp", strconv.FormatInt(now, 10))
	params.Set("expires_at", strconv.FormatInt(expiresAt, 10))

	signature, err := api.SignParametersUsingAlgoAndVersion(
		params,
		cld.Config.Cloud.APISecret,
		cld.Config.Cloud.GetSignatureAlgorithm(),
		cld.Config.Cloud.GetSignatureVersion(),
	)
	if err != nil {
		return "", err
	}

	params.Set("signature", signature)
	params.Set("api_key", cld.Config.Cloud.APIKey)

	baseURL := api.BaseURL(cld.Config.API.UploadPrefix, "")
	endpoint := fmt.Sprintf("%s/%s/%s", baseURL, cld.Config.Cloud.CloudName, api.BuildPath(assetRef.ResourceType, "download"))

	urlStruct, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	urlStruct.RawQuery = params.Encode()

	return urlStruct.String(), nil
}
