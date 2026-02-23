package api

import (
	"net/http"
	"strconv"

	"nagare/internal/model"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAllItemsCtrl handles GET /items
func GetAllItemsCtrl(c *gin.Context) {
	items, err := service.GetAllItemServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

// SearchItemsCtrl handles GET /item/search
func SearchItemsCtrl(c *gin.Context) {
	hid, err := parseOptionalUint(c, "hid")
	if err != nil {
		respondBadRequest(c, "invalid hid")
		return
	}
	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}
	withTotal, _ := parseOptionalBool(c, "with_total")
	limit := 100
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	offset := 0
	if o, err := parseOptionalInt(c, "offset"); err == nil && o != nil {
		offset = *o
	}
	filter := model.ItemFilter{
		Query:       c.Query("q"),
		SearchField: c.Query("search_field"),
		HID:         hid,
		ValueType:   parseOptionalString(c, "value_type"),
		Status:    status,
		HostID:    parseOptionalString(c, "hostid"),
		ItemID:    parseOptionalString(c, "itemid"),
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	items, err := service.SearchItemsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountItemsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": items, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

// GetItemByIDCtrl handles GET /items/:id
func GetItemByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid item ID")
		return
	}

	item, err := service.GetItemByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, item)
}

// GetItemHistoryCtrl handles GET /items/:id/history
func GetItemHistoryCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid item ID")
		return
	}
	from, err := parseOptionalUnixTime(c, "from")
	if err != nil {
		respondBadRequest(c, "invalid from timestamp")
		return
	}
	to, err := parseOptionalUnixTime(c, "to")
	if err != nil {
		respondBadRequest(c, "invalid to timestamp")
		return
	}
	limit := 500
	if l, err := parseOptionalInt(c, "limit"); err == nil && l != nil {
		limit = *l
	}
	items, err := service.GetItemHistoryServ(uint(id), from, to, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

// AddItemCtrl handles POST /item
func AddItemCtrl(c *gin.Context) {
	var req service.ItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	item, err := service.AddItemServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, item)
}

// UpdateItemCtrl handles PUT /item/:id
func UpdateItemCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid item ID")
		return
	}

	var req service.ItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateItemServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "item updated")
}

// DeleteItemByIDCtrl handles DELETE /item/:id
func DeleteItemByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid item ID")
		return
	}

	pushToMonitor := c.Query("push") == "true"
	if err := service.DeleteItemByIDServ(uint(id), pushToMonitor); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "item deleted")
}

// AddItemsByHostIDFromMonitorCtrl handles POST /items/by-host/:hid
func AddItemsByHostIDFromMonitorCtrl(c *gin.Context) {
	hid, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	if err := service.AddItemsByHostIDFromMonitorServ(uint(hid)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusCreated, "items added from monitor")
}

func PullItemsFromMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	result, err := service.PullItemsFromMonitorServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

func PullItemsOfHostFromMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	result, err := service.PullItemsFromHostServ(uint(mid), uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

func PullItemOfHostFromMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("item_id"))
	if err != nil {
		respondBadRequest(c, "invalid item ID")
		return
	}

	hid, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	result, err := service.PullItemOfHostFromMonitorServ(uint(mid), uint(hid), uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

func PushItemToMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("item_id"))
	if err != nil {
		respondBadRequest(c, "invalid item ID")
		return
	}

	hid, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	if err := service.PushItemToMonitorServ(uint(mid), uint(hid), uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "item pushed to monitor")
}

func PushItemsFromHostCtrl(c *gin.Context) {
	hid, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	result, err := service.PushItemsFromHostServ(uint(mid), uint(hid))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

func PushItemsFromMonitorCtrl(c *gin.Context) {
	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	result, err := service.PushItemsFromMonitorServ(uint(mid))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

// GenerateTestHistoryCtrl - DEVELOPMENT ONLY: Generates test history data for debugging charts
func GenerateTestHistoryCtrl(c *gin.Context) {
	if err := service.GenerateTestHistoryServ(); err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, gin.H{"message": "test history data generated"})
}
