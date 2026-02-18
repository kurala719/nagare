package api

import (
	"net/http"
	"strconv"

	"nagare/internal/model"
	"nagare/internal/service"

	"github.com/gin-gonic/gin"
)

// GetAllHostsCtrl handles GET /hosts
func GetAllHostsCtrl(c *gin.Context) {
	hosts, err := service.GetAllHostsServ()
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, hosts)
}

// SearchHostsCtrl handles GET /host/search
func SearchHostsCtrl(c *gin.Context) {
	status, err := parseOptionalInt(c, "status")
	if err != nil {
		respondBadRequest(c, "invalid status")
		return
	}
	mid, err := parseOptionalUint(c, "m_id")
	if err != nil {
		respondBadRequest(c, "invalid m_id")
		return
	}
	groupID, err := parseOptionalUint(c, "group_id")
	if err != nil {
		respondBadRequest(c, "invalid group_id")
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
	filter := model.HostFilter{
		Query:     c.Query("q"),
		MID:       mid,
		GroupID:   groupID,
		Status:    status,
		IPAddr:    parseOptionalString(c, "ip_addr"),
		Limit:     limit,
		Offset:    offset,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
	hosts, err := service.SearchHostsServ(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	if withTotal != nil && *withTotal {
		total, err := service.CountHostsServ(filter)
		if err != nil {
			respondError(c, err)
			return
		}
		respondSuccess(c, http.StatusOK, gin.H{"items": hosts, "total": total})
		return
	}
	respondSuccess(c, http.StatusOK, hosts)
}

// GetHostByIDCtrl handles GET /hosts/:id
func GetHostByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	host, err := service.GetHostByIDServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, host)
}

// GetHostHistoryCtrl handles GET /hosts/:id/history
func GetHostHistoryCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
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
	items, err := service.GetHostHistoryServ(uint(id), from, to, limit)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

// AddHostCtrl handles POST /host
func AddHostCtrl(c *gin.Context) {
	var req service.HostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	host, err := service.AddHostServ(req)
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusCreated, host)
}

// UpdateHostCtrl handles PUT /host/:id
func UpdateHostCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	var req service.HostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	if err := service.UpdateHostServ(uint(id), req); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "host updated")
}

// DeleteHostByIDCtrl handles DELETE /host/:id
func DeleteHostByIDCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	if err := service.DeleteHostByIDServ(uint(id)); err != nil {
		respondError(c, err)
		return
	}
	respondSuccessMessage(c, http.StatusOK, "host deleted")
}

// GetHostsFromMonitorCtrl handles GET /hosts/from-monitor/:id
func GetHostsFromMonitorCtrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	hosts, err := service.GetHostsFromMonitorServ(uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, hosts)
}

func PullHostsFromMonitorCtrl(c *gin.Context) {
	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	results, err := service.PullHostsFromMonitorServ(uint(mid))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, results)
}

func PullHostFromMonitorCtrl(c *gin.Context) {
	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	id, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	results, err := service.PullHostFromMonitorServ(uint(mid), uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, results)
}

func PushHostToMonitorCtrl(c *gin.Context) {
	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	id, err := strconv.Atoi(c.Param("hid"))
	if err != nil {
		respondBadRequest(c, "invalid host ID")
		return
	}

	result, err := service.PushHostToMonitorServ(uint(mid), uint(id))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, result)
}

func PushHostsFromMonitorCtrl(c *gin.Context) {
	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid monitor ID")
		return
	}

	results, err := service.PushHostsFromMonitorServ(uint(mid))
	if err != nil {
		respondError(c, err)
		return
	}
	respondSuccess(c, http.StatusOK, results)
}
