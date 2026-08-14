package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	mw "github.com/ericrabun/findfore-go/internal/adapter/inbound/http/middleware"
	"github.com/ericrabun/findfore-go/internal/application/apperr"
	"github.com/ericrabun/findfore-go/internal/application/groups"
	"github.com/ericrabun/findfore-go/internal/domain/entity"
	"github.com/ericrabun/findfore-go/internal/domain/port"
)

type groupViewerResponse struct {
	Status string `json:"status"`
	Role   string `json:"role"`
}

type groupResponse struct {
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Privacy     string               `json:"privacy"`
	Owner       map[string]any       `json:"owner"`
	MemberCount int64                `json:"member_count"`
	Viewer      *groupViewerResponse `json:"viewer_membership"`
}

type groupMemberResponse struct {
	PlayerID   int64  `json:"player_id"`
	PlayerName string `json:"player_name"`
	Role       string `json:"role"`
	Status     string `json:"status"`
}

type groupInvitationResponse struct {
	ID          int64  `json:"id"`
	GroupID     int64  `json:"group_id"`
	GroupName   string `json:"group_name,omitempty"`
	InviterID   int64  `json:"inviter_player_id"`
	InviterName string `json:"inviter_name,omitempty"`
	InviteeID   int64  `json:"invitee_player_id"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

type createGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Privacy     string `json:"privacy"`
}

type inviteGroupRequest struct {
	PlayerID int64 `json:"player_id"`
}

func mapGroupDetails(d *port.GroupDetails) groupResponse {
	out := groupResponse{
		ID:          d.Group.ID,
		Name:        d.Group.Name,
		Description: d.Group.Description,
		Privacy:     d.Group.Privacy,
		Owner:       map[string]any{"id": d.Group.OwnerPlayerID, "name": d.OwnerName},
		MemberCount: d.MemberCount,
	}
	if d.Viewer != nil {
		out.Viewer = &groupViewerResponse{Status: d.Viewer.Status, Role: d.Viewer.Role}
	}
	return out
}

func mapMembership(m *entity.GroupMembership) groupMemberResponse {
	return groupMemberResponse{
		PlayerID: m.PlayerID, Role: m.Role, Status: m.Status,
	}
}

func mapInvitation(inv *entity.GroupInvitation, groupName, inviterName string) groupInvitationResponse {
	out := groupInvitationResponse{
		ID: inv.ID, GroupID: inv.GroupID, GroupName: groupName,
		InviterID: inv.InviterPlayerID, InviterName: inviterName,
		InviteeID: inv.InviteePlayerID,
	}
	if inv.ExpiresAt != nil {
		out.ExpiresAt = inv.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return out
}

func parseLimitOffset(r *http.Request) (int32, int32) {
	limit, offset := int32(20), int32(0)
	if v, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32); err == nil {
		limit = int32(v)
	}
	if v, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32); err == nil {
		offset = int32(v)
	}
	return limit, offset
}

func parseIDParam(r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (h *Handler) requireGroups(w http.ResponseWriter) bool {
	if h.groups == nil {
		respondError(w, http.StatusNotImplemented, "not_implemented", "Groups is not configured")
		return false
	}
	return true
}

func writeGroupError(w http.ResponseWriter, r *http.Request, err error) {
	var ve *apperr.ValidationError
	switch {
	case errors.As(err, &ve):
		respondError(w, http.StatusBadRequest, "validation_error", ve.Message)
	case errors.Is(err, groups.ErrGroupNotFound), errors.Is(err, groups.ErrInvitationNotFound):
		respondError(w, http.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, groups.ErrGroupForbidden):
		respondError(w, http.StatusForbidden, "forbidden", "Not allowed to modify this group")
	case errors.Is(err, groups.ErrGroupConflict), errors.Is(err, entity.ErrGroupConflict):
		respondError(w, http.StatusConflict, "conflict", "Group relationship conflict")
	case errors.Is(err, groups.ErrGroupOwnerCannotLeave):
		respondError(w, http.StatusConflict, "conflict", "Group owner cannot leave")
	case errors.Is(err, groups.ErrInvitationExpired):
		respondError(w, http.StatusConflict, "conflict", "Invitation expired")
	case errors.Is(err, groups.ErrInvalidGroup):
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group request")
	default:
		respondInternalError(w, r, err, "Group request failed")
	}
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid JSON body")
		return
	}
	d, err := h.groups.Create(r.Context(), port.CreateGroupInput{
		ActorID: actorID, Name: req.Name, Description: req.Description, Privacy: req.Privacy,
	})
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusCreated, mapGroupDetails(d))
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	d, err := h.groups.Get(r.Context(), actorID, id)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, mapGroupDetails(d))
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	limit, offset := parseLimitOffset(r)
	var (
		list []port.GroupDetails
		err  error
	)
	if r.URL.Query().Get("mine") == "1" || r.URL.Query().Get("mine") == "true" {
		list, err = h.groups.ListMine(r.Context(), actorID, limit, offset)
	} else {
		list, err = h.groups.ListDiscover(r.Context(), actorID, r.URL.Query().Get("search"), limit, offset)
	}
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	items := make([]groupResponse, len(list))
	for i := range list {
		items[i] = mapGroupDetails(&list[i])
	}
	respondJSON(w, http.StatusOK, map[string]any{"groups": items})
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid JSON body")
		return
	}
	d, err := h.groups.Update(r.Context(), port.UpdateGroupInput{
		ActorID: actorID, GroupID: id, Name: req.Name, Description: req.Description, Privacy: req.Privacy,
	})
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, mapGroupDetails(d))
}

func (h *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	m, err := h.groups.Join(r.Context(), actorID, id)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, mapMembership(m))
}

func (h *Handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	if err := h.groups.Leave(r.Context(), actorID, id); err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	limit, offset := parseLimitOffset(r)
	members, err := h.groups.ListMembers(r.Context(), actorID, id, limit, offset)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	items := make([]groupMemberResponse, len(members))
	for i, m := range members {
		items[i] = groupMemberResponse{PlayerID: m.PlayerID, PlayerName: m.PlayerName, Role: m.Role, Status: m.Status}
	}
	respondJSON(w, http.StatusOK, map[string]any{"members": items})
}

func (h *Handler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	playerID, ok := parseIDParam(r, "playerId")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid player id")
		return
	}
	if err := h.groups.RemoveMember(r.Context(), actorID, id, playerID); err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) InviteToGroup(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	var req inviteGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerID <= 0 {
		respondError(w, http.StatusBadRequest, "validation_error", "player_id is required")
		return
	}
	inv, err := h.groups.Invite(r.Context(), actorID, id, req.PlayerID)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	if inv == nil {
		respondJSON(w, http.StatusOK, map[string]any{"status": "active"})
		return
	}
	respondJSON(w, http.StatusCreated, mapInvitation(inv, "", ""))
}

func (h *Handler) ListJoinRequests(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	rows, err := h.groups.ListJoinRequests(r.Context(), actorID, id)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	items := make([]groupMemberResponse, len(rows))
	for i, m := range rows {
		items[i] = groupMemberResponse{PlayerID: m.PlayerID, PlayerName: m.PlayerName, Role: m.Role, Status: m.Status}
	}
	respondJSON(w, http.StatusOK, map[string]any{"join_requests": items})
}

func (h *Handler) ApproveJoinRequest(w http.ResponseWriter, r *http.Request) {
	h.mutateJoinRequest(w, r, true)
}

func (h *Handler) DenyJoinRequest(w http.ResponseWriter, r *http.Request) {
	h.mutateJoinRequest(w, r, false)
}

func (h *Handler) mutateJoinRequest(w http.ResponseWriter, r *http.Request, approve bool) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid group id")
		return
	}
	playerID, ok := parseIDParam(r, "playerId")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid player id")
		return
	}
	if approve {
		m, err := h.groups.ApproveJoinRequest(r.Context(), actorID, id, playerID)
		if err != nil {
			writeGroupError(w, r, err)
			return
		}
		respondJSON(w, http.StatusOK, mapMembership(m))
		return
	}
	if err := h.groups.DenyJoinRequest(r.Context(), actorID, id, playerID); err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) ListMyGroupInvitations(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	rows, err := h.groups.ListMyInvitations(r.Context(), actorID)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	items := make([]groupInvitationResponse, len(rows))
	for i, row := range rows {
		items[i] = mapInvitation(&row.Invitation, row.GroupName, row.InviterName)
	}
	respondJSON(w, http.StatusOK, map[string]any{"invitations": items})
}

func (h *Handler) AcceptGroupInvitation(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid invitation id")
		return
	}
	m, err := h.groups.AcceptInvitation(r.Context(), actorID, id)
	if err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, mapMembership(m))
}

func (h *Handler) DeclineGroupInvitation(w http.ResponseWriter, r *http.Request) {
	if !h.requireGroups(w) {
		return
	}
	actorID, ok := mw.PlayerIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}
	id, ok := parseIDParam(r, "id")
	if !ok {
		respondError(w, http.StatusBadRequest, "validation_error", "Invalid invitation id")
		return
	}
	if err := h.groups.DeclineInvitation(r.Context(), actorID, id); err != nil {
		writeGroupError(w, r, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}
