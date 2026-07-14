package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"vault_api/internal/service"
)

func parseVaultListFilter(r *http.Request) (service.ListVaultItemsFilter, error) {
	query := r.URL.Query()

	filter := service.ListVaultItemsFilter{
		Folder:   strings.TrimSpace(query.Get("folder")),
		ItemType: strings.TrimSpace(query.Get("item_type")),
		Tag:      strings.TrimSpace(query.Get("tag")),
		Title:    strings.TrimSpace(query.Get("title")),
	}

	if v := query.Get("limit"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed <= 0 {
			return service.ListVaultItemsFilter{}, fmt.Errorf("invalid limit")
		}
		filter.Limit = int32(parsed)
	}

	if v := query.Get("offset"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil || parsed < 0 {
			return service.ListVaultItemsFilter{}, fmt.Errorf("invalid offset")
		}
		filter.Offset = int32(parsed)
	}

	return filter, nil
}
