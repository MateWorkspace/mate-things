package presentationhttphandleradmin

import (
	"net/http"

	domainusecasesadmin "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/admin"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type handler struct {
	permissionUseCase domainusecasesadmin.PermissionManagement
	roleUseCase       domainusecasesadmin.RoleManagement
	schemaUseCase     domainusecasesadmin.SchemaRegistry
	userUseCase       domainusecasesadmin.UserManagement
}

func NewHandler(
	permissionUseCase domainusecasesadmin.PermissionManagement,
	roleUseCase domainusecasesadmin.RoleManagement,
	schemaUseCase domainusecasesadmin.SchemaRegistry,
	userUseCase domainusecasesadmin.UserManagement,
) *handler {
	return &handler{
		permissionUseCase: permissionUseCase,
		roleUseCase:       roleUseCase,
		schemaUseCase:     schemaUseCase,
		userUseCase:       userUseCase,
	}
}

// PermissionPost godoc
//
// @Summary Permission
// @Tags Admin - Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.PermissionPostRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/permissions [post]
func (h *handler) PermissionPost(c *echo.Context) error {
	var req presentationhttprequest.PermissionPostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	id, err := h.permissionUseCase.Create(c.Request().Context(), domainusecasesadmin.CreatePermissionRequest{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to create the permission. Please check your input and try again.")
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// PermissionGetList godoc
//
// @Summary Permission List
// @Tags Admin - Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.PermissionResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/permissions [get]
func (h *handler) PermissionGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}

	permissions, total, err := h.permissionUseCase.ReadByPagination(c.Request().Context(), domainusecasesadmin.ReadPermissionsByPaginationRequest{
		Page:   page.Page,
		Limit:  page.Limit,
		Search: page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load permissions right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.PermissionResponse]{
		Data: presentationhttpresponse.Permissions(permissions),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// PermissionGetByName godoc
//
// @Summary Permission Get By Name
// @Tags Admin - Permissions
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200 {object} presentationhttpresponse.PermissionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/permissions/by-name/{name} [get]
func (h *handler) PermissionGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid permission name.")
	}

	permission, err := h.permissionUseCase.ReadByName(c.Request().Context(), domainusecasesadmin.ReadPermissionByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the permission. Please try again.")
	}
	if permission == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("permission"), "The requested permission could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Permission(*permission))
}

// PermissionGetById godoc
//
// @Summary Permission Get By ID
// @Tags Admin - Permissions
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.PermissionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/permissions/{id} [get]
func (h *handler) PermissionGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The permission ID provided is invalid.")
	}

	permission, err := h.permissionUseCase.ReadById(c.Request().Context(), domainusecasesadmin.ReadPermissionByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the permission. Please try again.")
	}
	if permission == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("permission"), "The requested permission could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Permission(*permission))
}

// PermissionPatch godoc
//
// @Summary Permission
// @Tags Admin - Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.PermissionPatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/permissions/{id} [patch]
func (h *handler) PermissionPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The permission ID provided is invalid.")
	}

	var req presentationhttprequest.PermissionPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.permissionUseCase.UpdateById(c.Request().Context(), domainusecasesadmin.UpdatePermissionRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to update the permission. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// PermissionDelete godoc
//
// @Summary Permission Delete
// @Tags Admin - Permissions
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/permissions/{id} [delete]
func (h *handler) PermissionDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The permission ID provided is invalid.")
	}

	if err := h.permissionUseCase.DeleteById(c.Request().Context(), domainusecasesadmin.DeletePermissionRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete the permission. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// RolePost godoc
//
// @Summary Role
// @Tags Admin - Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.RolePostRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles [post]
func (h *handler) RolePost(c *echo.Context) error {
	var req presentationhttprequest.RolePostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	id, err := h.roleUseCase.Create(c.Request().Context(), domainusecasesadmin.CreateRoleRequest{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to create the role. Please check your input and try again.")
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// RoleGetList godoc
//
// @Summary Role List
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.RoleResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles [get]
func (h *handler) RoleGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}

	roles, total, err := h.roleUseCase.ReadByPagination(c.Request().Context(), domainusecasesadmin.ReadRolesByPaginationRequest{
		Page:   page.Page,
		Limit:  page.Limit,
		Search: page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load roles right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.RoleResponse]{
		Data: presentationhttpresponse.Roles(roles),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// RoleGetDefault godoc
//
// @Summary Role Get Default
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.RoleResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/default [get]
func (h *handler) RoleGetDefault(c *echo.Context) error {
	role, err := h.roleUseCase.ReadDefault(c.Request().Context(), domainusecasesadmin.ReadDefaultRoleRequest{})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the default role. Please try again.")
	}
	if role == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("role"), "No default role has been set.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Role(*role))
}

// RoleGetByName godoc
//
// @Summary Role Get By Name
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200 {object} presentationhttpresponse.RoleResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/by-name/{name} [get]
func (h *handler) RoleGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid role name.")
	}

	role, err := h.roleUseCase.ReadByName(c.Request().Context(), domainusecasesadmin.ReadRoleByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the role. Please try again.")
	}
	if role == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("role"), "The requested role could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Role(*role))
}

// RoleGetById godoc
//
// @Summary Role Get By ID
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.RoleResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/{id} [get]
func (h *handler) RoleGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}

	role, err := h.roleUseCase.ReadById(c.Request().Context(), domainusecasesadmin.ReadRoleByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the role. Please try again.")
	}
	if role == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("role"), "The requested role could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Role(*role))
}

// RolePermissionsGet godoc
//
// @Summary Role Permissions Get
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.PermissionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/{id}/permissions [get]
func (h *handler) RolePermissionsGet(c *echo.Context) error {
	roleId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}

	permissions, err := h.roleUseCase.ReadPermissions(c.Request().Context(), domainusecasesadmin.ReadRolePermissionsRequest{RoleId: roleId})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load permissions for this role right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Permissions(permissions))
}

// RolePatch godoc
//
// @Summary Role
// @Tags Admin - Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.RolePatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/{id} [patch]
func (h *handler) RolePatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}

	var req presentationhttprequest.RolePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.roleUseCase.UpdateById(c.Request().Context(), domainusecasesadmin.UpdateRoleRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to update the role. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// RoleSetDefaultPatch godoc
//
// @Summary Role Set Default
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/{id}/default [patch]
func (h *handler) RoleSetDefaultPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}

	if err := h.roleUseCase.SetDefaultRole(c.Request().Context(), domainusecasesadmin.SetDefaultRoleRequest{
		Id:        id,
		UpdatedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to set the default role right now. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// RoleDelete godoc
//
// @Summary Role Delete
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/{id} [delete]
func (h *handler) RoleDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}

	if err := h.roleUseCase.DeleteById(c.Request().Context(), domainusecasesadmin.DeleteRoleRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete the role. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// RolePermissionPost godoc
//
// @Summary Role Permission
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Param role_id path string true "role_id"
// @Param permission_id path string true "permission_id"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/{role_id}/permissions/{permission_id} [post]
func (h *handler) RolePermissionPost(c *echo.Context) error {
	roleId, permissionId, err := h.rolePermissionPath(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The role or permission ID provided is invalid.")
	}

	id, err := h.roleUseCase.AssignPermission(c.Request().Context(), domainusecasesadmin.AssignRolePermissionRequest{
		RoleId:       roleId,
		PermissionId: permissionId,
		CreatedBy:    presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to assign this permission to the role. Please try again.")
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// RolePermissionDeleteByPair godoc
//
// @Summary Role Permission Delete By Pair
// @Tags Admin - Roles
// @Produce json
// @Security BearerAuth
// @Param role_id path string true "role_id"
// @Param permission_id path string true "permission_id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/roles/{role_id}/permissions/{permission_id} [delete]
func (h *handler) RolePermissionDeleteByPair(c *echo.Context) error {
	roleId, permissionId, err := h.rolePermissionPath(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The role or permission ID provided is invalid.")
	}

	if err := h.roleUseCase.RevokePermission(c.Request().Context(), domainusecasesadmin.RevokeRolePermissionRequest{
		RoleId:       roleId,
		PermissionId: permissionId,
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to remove this permission from the role. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// RolePermissionGetList godoc
//
// @Summary Role Permission List
// @Tags Admin - Role Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.RolePermissionDetailResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/role-permissions [get]
func (h *handler) RolePermissionGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}
	roleId, err := presentationhttputils.QueryUUID(c, "role_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}
	permissionId, err := presentationhttputils.QueryUUID(c, "permission_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The permission ID provided is invalid.")
	}

	rolePermissions, roles, permissions, total, err := h.roleUseCase.ReadRolePermissionsByPagination(
		c.Request().Context(),
		domainusecasesadmin.ReadRolePermissionsByPaginationRequest{
			Page:         page.Page,
			Limit:        page.Limit,
			RoleId:       roleId,
			PermissionId: permissionId,
		},
	)
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load role permissions right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.RolePermissionDetailResponse]{
		Data: presentationhttpresponse.RolePermissionDetails(rolePermissions, roles, permissions),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// RolePermissionGetById godoc
//
// @Summary Role Permission Get By ID
// @Tags Admin - Role Permissions
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.RolePermissionDetailResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/role-permissions/{id} [get]
func (h *handler) RolePermissionGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role permission ID provided is invalid.")
	}

	result, err := h.roleUseCase.ReadRolePermissionById(c.Request().Context(), domainusecasesadmin.ReadRolePermissionByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "The requested role permission could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.RolePermissionDetail(result.RolePermission, result.Role, result.Permission))
}

// RolePermissionGetByPair godoc
//
// @Summary Role Permission Get By Pair
// @Tags Admin - Role Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.RolePermissionDetailResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/role-permissions/by-pair [get]
func (h *handler) RolePermissionGetByPair(c *echo.Context) error {
	roleId, err := presentationhttputils.QueryUUID(c, "role_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}
	permissionId, err := presentationhttputils.QueryUUID(c, "permission_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The permission ID provided is invalid.")
	}
	if roleId == nil || permissionId == nil {
		return presentationhttputils.Error(c, presentationhttputils.BadPairQuery("role_id", "permission_id"), "Both role_id and permission_id are required.")
	}

	result, err := h.roleUseCase.ReadRolePermissionByRoleIdAndPermissionId(
		c.Request().Context(),
		domainusecasesadmin.ReadRolePermissionByRoleIdAndPermissionIdRequest{RoleId: *roleId, PermissionId: *permissionId},
	)
	if err != nil {
		return presentationhttputils.Error(c, err, "The requested role permission could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.RolePermissionDetail(result.RolePermission, result.Role, result.Permission))
}

func (h *handler) rolePermissionPath(c *echo.Context) (uuid.UUID, uuid.UUID, error) {
	roleId, err := presentationhttputils.RequiredUUID(c.Param("role_id"), "role_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	permissionId, err := presentationhttputils.RequiredUUID(c.Param("permission_id"), "permission_id")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return roleId, permissionId, nil
}

// PayloadSchemaPost godoc
//
// @Summary Payload Schema
// @Tags Admin - Payload Schemas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.PayloadSchemaPostRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/payload-schemas [post]
func (h *handler) PayloadSchemaPost(c *echo.Context) error {
	var req presentationhttprequest.PayloadSchemaPostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	definition, err := presentationhttputils.RequiredRawJSON(req.Definition, "definition")
	if err != nil {
		return presentationhttputils.Error(c, err, "A valid schema definition is required.")
	}

	id, err := h.schemaUseCase.Create(c.Request().Context(), domainusecasesadmin.CreatePayloadSchemaRequest{
		Name:       req.Name,
		Version:    req.Version,
		Definition: definition,
		ValidFrom:  req.ValidFrom,
		ValidTo:    req.ValidTo,
		CreatedBy:  presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to create the payload schema. Please check your input and try again.")
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// PayloadSchemaGetList godoc
//
// @Summary Payload Schema List
// @Tags Admin - Payload Schemas
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.PayloadSchemaResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/payload-schemas [get]
func (h *handler) PayloadSchemaGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}
	validAt, err := presentationhttputils.QueryTime(c, "valid_at")
	if err != nil {
		return presentationhttputils.Error(c, err, "The valid_at date provided is invalid.")
	}

	payloadSchemas, total, err := h.schemaUseCase.ReadByPagination(c.Request().Context(), domainusecasesadmin.ReadPayloadSchemasByPaginationRequest{
		Page:    page.Page,
		Limit:   page.Limit,
		Search:  page.Search,
		ValidAt: validAt,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load payload schemas right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.PayloadSchemaResponse]{
		Data: presentationhttpresponse.PayloadSchemas(payloadSchemas),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// PayloadSchemaGetLatest godoc
//
// @Summary Payload Schema Get Latest
// @Tags Admin - Payload Schemas
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PayloadSchemaResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/payload-schemas/latest [get]
func (h *handler) PayloadSchemaGetLatest(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.QueryParam("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid payload schema name.")
	}

	payloadSchema, err := h.schemaUseCase.ReadLatestByName(c.Request().Context(), domainusecasesadmin.ReadLatestPayloadSchemaByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the payload schema. Please try again.")
	}
	if payloadSchema == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("payload schema"), "The requested payload schema could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PayloadSchema(*payloadSchema))
}

// PayloadSchemaGetByNameAndVersion godoc
//
// @Summary Payload Schema Get By Name And Version
// @Tags Admin - Payload Schemas
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PayloadSchemaResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/payload-schemas/by-name-version [get]
func (h *handler) PayloadSchemaGetByNameAndVersion(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.QueryParam("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid payload schema name.")
	}
	version, err := presentationhttputils.RequiredInt32(c.QueryParam("version"), "version")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid payload schema version.")
	}

	payloadSchema, err := h.schemaUseCase.ReadByNameAndVersion(c.Request().Context(), domainusecasesadmin.ReadPayloadSchemaByNameAndVersionRequest{
		Name:    name,
		Version: version,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the payload schema. Please try again.")
	}
	if payloadSchema == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("payload schema"), "The requested payload schema could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PayloadSchema(*payloadSchema))
}

// PayloadSchemaGetById godoc
//
// @Summary Payload Schema Get By ID
// @Tags Admin - Payload Schemas
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.PayloadSchemaResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/payload-schemas/{id} [get]
func (h *handler) PayloadSchemaGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The payload schema ID provided is invalid.")
	}

	payloadSchema, err := h.schemaUseCase.ReadById(c.Request().Context(), domainusecasesadmin.ReadPayloadSchemaByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the payload schema. Please try again.")
	}
	if payloadSchema == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("payload schema"), "The requested payload schema could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PayloadSchema(*payloadSchema))
}

// PayloadSchemaPatch godoc
//
// @Summary Payload Schema
// @Tags Admin - Payload Schemas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.PayloadSchemaPatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/payload-schemas/{id} [patch]
func (h *handler) PayloadSchemaPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The payload schema ID provided is invalid.")
	}

	var req presentationhttprequest.PayloadSchemaPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	definition, err := presentationhttputils.OptionalRawJSON(req.Definition, "definition")
	if err != nil {
		return presentationhttputils.Error(c, err, "The schema definition provided is invalid.")
	}

	if err := h.schemaUseCase.UpdateById(c.Request().Context(), domainusecasesadmin.UpdatePayloadSchemaRequest{
		Id:         id,
		Name:       req.Name,
		Version:    req.Version,
		Definition: definition,
		ValidFrom:  req.ValidFrom,
		ValidTo:    req.ValidTo,
		UpdatedBy:  presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to update the payload schema. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// PayloadSchemaDelete godoc
//
// @Summary Payload Schema Delete
// @Tags Admin - Payload Schemas
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/payload-schemas/{id} [delete]
func (h *handler) PayloadSchemaDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The payload schema ID provided is invalid.")
	}

	if err := h.schemaUseCase.DeleteById(c.Request().Context(), domainusecasesadmin.DeletePayloadSchemaRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete the payload schema. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// UserPost godoc
//
// @Summary User
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.UserPostRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users [post]
func (h *handler) UserPost(c *echo.Context) error {
	var req presentationhttprequest.UserPostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	roleId, err := presentationhttputils.RequiredUUID(req.RoleId, "role_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please select a valid role.")
	}

	id, err := h.userUseCase.Create(c.Request().Context(), domainusecasesadmin.CreateUserRequest{
		RoleId:    roleId,
		Name:      req.Name,
		Bio:       req.Bio,
		Username:  req.Username,
		Password:  req.Password,
		CreatedBy: presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to create the user. Please check your input and try again.")
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// UserGetList godoc
//
// @Summary User List
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.UserResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users [get]
func (h *handler) UserGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}
	roleId, err := presentationhttputils.QueryUUID(c, "role_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}

	users, total, err := h.userUseCase.ReadByPagination(c.Request().Context(), domainusecasesadmin.ReadUsersByPaginationRequest{
		Page:   page.Page,
		Limit:  page.Limit,
		Search: page.Search,
		RoleId: roleId,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load users right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.UserResponse]{
		Data: presentationhttpresponse.Users(users),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// UserGetByUsername godoc
//
// @Summary User Get By Username
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param username path string true "username"
// @Success 200 {object} presentationhttpresponse.UserResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users/by-username/{username} [get]
func (h *handler) UserGetByUsername(c *echo.Context) error {
	username, err := presentationhttputils.RequiredString(c.Param("username"), "username")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid username.")
	}

	user, err := h.userUseCase.ReadByUsername(c.Request().Context(), domainusecasesadmin.ReadUserByUsernameRequest{Username: username})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the user. Please try again.")
	}
	if user == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("user"), "The requested user could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.User(*user))
}

// UserGetById godoc
//
// @Summary User Get By ID
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.UserResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users/{id} [get]
func (h *handler) UserGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The user ID provided is invalid.")
	}

	user, err := h.userUseCase.ReadById(c.Request().Context(), domainusecasesadmin.ReadUserByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the user. Please try again.")
	}
	if user == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("user"), "The requested user could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.User(*user))
}

// UserPermissionsGet godoc
//
// @Summary User Permissions Get
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.PermissionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users/{id}/permissions [get]
func (h *handler) UserPermissionsGet(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The user ID provided is invalid.")
	}

	permissions, err := h.userUseCase.ReadPermissions(c.Request().Context(), domainusecasesadmin.ReadUserPermissionsRequest{UserId: userId})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load permissions for this user right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Permissions(permissions))
}

// UserPatch godoc
//
// @Summary User
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.UserPatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users/{id} [patch]
func (h *handler) UserPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The user ID provided is invalid.")
	}

	var req presentationhttprequest.UserPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	roleId, err := presentationhttputils.OptionalUUID(req.RoleId, "role_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The role ID provided is invalid.")
	}

	if err := h.userUseCase.UpdateById(c.Request().Context(), domainusecasesadmin.UpdateUserRequest{
		Id:        id,
		RoleId:    roleId,
		Name:      req.Name,
		Bio:       req.Bio,
		Username:  req.Username,
		UpdatedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to update the user. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// UserPasswordPatch godoc
//
// @Summary User Password
// @Tags Admin - Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.UserPasswordPatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users/{id}/password [patch]
func (h *handler) UserPasswordPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The user ID provided is invalid.")
	}

	var req presentationhttprequest.UserPasswordPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.userUseCase.ResetPassword(c.Request().Context(), domainusecasesadmin.ResetUserPasswordRequest{
		Id:        id,
		Password:  req.Password,
		UpdatedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to reset the password. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// UserDelete godoc
//
// @Summary User Delete
// @Tags Admin - Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/admin/users/{id} [delete]
func (h *handler) UserDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The user ID provided is invalid.")
	}

	if err := h.userUseCase.DeleteById(c.Request().Context(), domainusecasesadmin.DeleteUserRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete the user. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}
