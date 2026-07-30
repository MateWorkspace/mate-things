package presentationhttphandlernode

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type handler struct {
	classUseCase           domainusecasesnode.ClassManagement
	deviceUseCase          domainusecasesnode.DeviceManagement
	firmwareUseCase        domainusecasesnode.FirmwareManagement
	otaUseCase             domainusecasesnode.Ota
	configParameterUseCase domainusecasesnode.ConfigParameter
	configValueUseCase     domainusecasesnode.ConfigValue
}

func NewHandler(
	classUseCase domainusecasesnode.ClassManagement,
	deviceUseCase domainusecasesnode.DeviceManagement,
	firmwareUseCase domainusecasesnode.FirmwareManagement,
	otaUseCase domainusecasesnode.Ota,
	configParameterUseCase domainusecasesnode.ConfigParameter,
	configValueUseCase domainusecasesnode.ConfigValue,
) *handler {
	return &handler{
		classUseCase:           classUseCase,
		deviceUseCase:          deviceUseCase,
		firmwareUseCase:        firmwareUseCase,
		otaUseCase:             otaUseCase,
		configParameterUseCase: configParameterUseCase,
		configValueUseCase:     configValueUseCase,
	}
}

// NodeClassPost godoc
//
// @Summary Node Class
// @Tags Node Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.NodeClassPostRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes [post]
func (h *handler) NodeClassPost(c *echo.Context) error {
	var req presentationhttprequest.NodeClassPostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	id, err := h.classUseCase.Create(c.Request().Context(), domainusecasesnode.CreateNodeClassRequest{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to create the node class. Please check your input and try again.")
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// NodeClassGetList godoc
//
// @Summary Node Class List
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.NodeClassResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes [get]
func (h *handler) NodeClassGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}

	nodeClasses, total, err := h.classUseCase.ReadByPagination(c.Request().Context(), domainusecasesnode.ReadNodeClassesByPaginationRequest{
		Page:   page.Page,
		Limit:  page.Limit,
		Search: page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load node classes right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.NodeClassResponse]{
		Data: presentationhttpresponse.NodeClasses(nodeClasses),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// NodeClassGetByName godoc
//
// @Summary Node Class Get By Name
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200 {object} presentationhttpresponse.NodeClassResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/by-name/{name} [get]
func (h *handler) NodeClassGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid node class name.")
	}

	nodeClass, err := h.classUseCase.ReadByName(c.Request().Context(), domainusecasesnode.ReadNodeClassByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the node class. Please try again.")
	}
	if nodeClass == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node class"), "The requested node class could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.NodeClass(*nodeClass))
}

// NodeClassGetById godoc
//
// @Summary Node Class Get By ID
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.NodeClassResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/{id} [get]
func (h *handler) NodeClassGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}

	nodeClass, err := h.classUseCase.ReadById(c.Request().Context(), domainusecasesnode.ReadNodeClassByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the node class. Please try again.")
	}
	if nodeClass == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node class"), "The requested node class could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.NodeClass(*nodeClass))
}

// NodeClassPatch godoc
//
// @Summary Node Class
// @Tags Node Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.NodeClassPatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/{id} [patch]
func (h *handler) NodeClassPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}

	var req presentationhttprequest.NodeClassPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.classUseCase.UpdateById(c.Request().Context(), domainusecasesnode.UpdateNodeClassRequest{
		Id:          id,
		Name:        req.Name,
		Description: req.Description,
		UpdatedBy:   presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to update the node class. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// NodeClassDelete godoc
//
// @Summary Node Class Delete
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/{id} [delete]
func (h *handler) NodeClassDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}

	if err := h.classUseCase.DeleteById(c.Request().Context(), domainusecasesnode.DeleteNodeClassRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete the node class. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// NodeGetList godoc
//
// @Summary Node List
// @Tags Nodes
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.NodeResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes [get]
func (h *handler) NodeGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}
	nodeClassId, err := presentationhttputils.QueryUUID(c, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}
	firmwareId, err := presentationhttputils.QueryUUID(c, "firmware_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}

	nodes, total, err := h.deviceUseCase.ReadByPagination(c.Request().Context(), domainusecasesnode.ReadNodesByPaginationRequest{
		Page:        page.Page,
		Limit:       page.Limit,
		Search:      page.Search,
		NodeClassId: nodeClassId,
		FirmwareId:  firmwareId,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load nodes right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.NodeResponse]{
		Data: presentationhttpresponse.Nodes(nodes),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// NodeGetByDeviceId godoc
//
// @Summary Node Get By Device ID
// @Tags Nodes
// @Produce json
// @Security BearerAuth
// @Param device_id path string true "device_id"
// @Success 200 {object} presentationhttpresponse.NodeResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/by-device/{device_id} [get]
func (h *handler) NodeGetByDeviceId(c *echo.Context) error {
	deviceId, err := presentationhttputils.RequiredString(c.Param("device_id"), "device_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid device ID.")
	}

	node, err := h.deviceUseCase.ReadByDeviceId(c.Request().Context(), domainusecasesnode.ReadNodeByDeviceIdRequest{DeviceId: deviceId})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the node. Please try again.")
	}
	if node == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node"), "The requested node could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Node(*node))
}

// NodeGetById godoc
//
// @Summary Node Get By ID
// @Tags Nodes
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.NodeResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/{id} [get]
func (h *handler) NodeGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}

	node, err := h.deviceUseCase.ReadById(c.Request().Context(), domainusecasesnode.ReadNodeByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the node. Please try again.")
	}
	if node == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node"), "The requested node could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Node(*node))
}

// NodePatch godoc
//
// @Summary Node
// @Tags Nodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.NodePatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/{id} [patch]
func (h *handler) NodePatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}

	var req presentationhttprequest.NodePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	nodeClassId, err := presentationhttputils.OptionalUUID(req.NodeClassId, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}
	firmwareId, err := presentationhttputils.OptionalUUID(req.FirmwareId, "firmware_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}

	if err := h.deviceUseCase.UpdateById(c.Request().Context(), domainusecasesnode.UpdateNodeRequest{
		Id:          id,
		NodeClassId: nodeClassId,
		Name:        req.Name,
		FirmwareId:  firmwareId,
		Description: req.Description,
		UpdatedBy:   presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to update the node. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// NodeFirmwarePatch godoc
//
// @Summary Node Firmware
// @Tags Nodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.NodeFirmwarePatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/{id}/firmware [patch]
func (h *handler) NodeFirmwarePatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}

	var req presentationhttprequest.NodeFirmwarePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	firmwareId, err := presentationhttputils.RequiredUUID(req.FirmwareId, "firmware_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please select a valid firmware.")
	}

	if err := h.deviceUseCase.AssignFirmware(c.Request().Context(), domainusecasesnode.AssignNodeFirmwareRequest{
		Id:         id,
		FirmwareId: firmwareId,
		UpdatedBy:  presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to assign this firmware to the node. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// NodeDelete godoc
//
// @Summary Node Delete
// @Tags Nodes
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/{id} [delete]
func (h *handler) NodeDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}

	if err := h.deviceUseCase.DeleteById(c.Request().Context(), domainusecasesnode.DeleteNodeRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete the node. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// FirmwarePost godoc
//
// @Summary Firmware
// @Tags Firmwares
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param node_class_id formData string true "node_class_id"
// @Param name formData string true "name"
// @Param file formData file true "file"
// @Param config_schema formData string false "config_schema"
// @Success 201 {object} presentationhttpresponse.FirmwareCreateResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares [post]
func (h *handler) FirmwarePost(c *echo.Context) error {
	nodeClassId, err := presentationhttputils.RequiredUUID(c.FormValue("node_class_id"), "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please select a valid node class.")
	}
	file, err := presentationhttputils.RequiredFormFile(c, "file")
	if err != nil {
		return presentationhttputils.Error(c, err, "A firmware file is required.")
	}

	content, err := file.Open()
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to read the uploaded file. Please try again.")
	}
	defer content.Close()

	configSchema, err := parseConfigSchemaFormValue(c.FormValue("config_schema"))
	if err != nil {
		return presentationhttputils.Error(c, err, "The config_schema field must be a valid JSON array of {key,value_type} objects.")
	}

	result, err := h.firmwareUseCase.Create(c.Request().Context(), domainusecasesnode.CreateFirmwareRequest{
		NodeClassId:  nodeClassId,
		Name:         c.FormValue("name"),
		Content:      content,
		ConfigSchema: configSchema,
		CreatedBy:    presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to upload the firmware. Please check your input and try again.")
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.FirmwareCreateResponse{
		Id:         result.Id.String(),
		Size:       result.Size,
		Checksum:   result.Checksum,
		BinaryPath: result.BinaryPath,
	})
}

// FirmwareGetList godoc
//
// @Summary Firmware List
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.FirmwareResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares [get]
func (h *handler) FirmwareGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}
	nodeClassId, err := presentationhttputils.QueryUUID(c, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}

	firmwares, total, err := h.firmwareUseCase.ReadByPagination(c.Request().Context(), domainusecasesnode.ReadFirmwaresByPaginationRequest{
		Page:        page.Page,
		Limit:       page.Limit,
		Search:      page.Search,
		NodeClassId: nodeClassId,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load firmwares right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.FirmwareResponse]{
		Data: presentationhttpresponse.Firmwares(firmwares),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// FirmwareGetByNodeClassId godoc
//
// @Summary Firmware Get By Node Class ID
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param node_class_id path string true "node_class_id"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.FirmwareResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-classes/{node_class_id}/firmwares [get]
func (h *handler) FirmwareGetByNodeClassId(c *echo.Context) error {
	nodeClassId, err := presentationhttputils.RequiredUUID(c.Param("node_class_id"), "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}

	firmwares, total, err := h.firmwareUseCase.ReadByNodeClassIdAndPagination(c.Request().Context(), domainusecasesnode.ReadFirmwaresByNodeClassIdAndPaginationRequest{
		NodeClassId: nodeClassId,
		Page:        page.Page,
		Limit:       page.Limit,
		Search:      page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load firmwares right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.FirmwareResponse]{
		Data: presentationhttpresponse.Firmwares(firmwares),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// FirmwareGetAvailableByNodeId godoc
//
// @Summary Firmware Get Available By Node ID
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param node_id path string true "node_id"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.FirmwareResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/{node_id}/firmwares/available [get]
func (h *handler) FirmwareGetAvailableByNodeId(c *echo.Context) error {
	nodeId, err := presentationhttputils.RequiredUUID(c.Param("node_id"), "node_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "The page or limit provided is invalid.")
	}

	firmwares, total, err := h.firmwareUseCase.ReadAvailableByNodeId(c.Request().Context(), domainusecasesnode.ReadAvailableFirmwaresByNodeIdRequest{
		NodeId: nodeId,
		Page:   page.Page,
		Limit:  page.Limit,
		Search: page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load available firmwares right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.FirmwareResponse]{
		Data: presentationhttpresponse.Firmwares(firmwares),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// FirmwareGetByName godoc
//
// @Summary Firmware Get By Name
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200 {object} presentationhttpresponse.FirmwareResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/by-name/{name} [get]
func (h *handler) FirmwareGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid firmware name.")
	}

	firmware, err := h.firmwareUseCase.ReadByName(c.Request().Context(), domainusecasesnode.ReadFirmwareByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the firmware. Please try again.")
	}
	if firmware == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("firmware"), "The requested firmware could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Firmware(*firmware))
}

// FirmwareGetById godoc
//
// @Summary Firmware Get By ID
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.FirmwareResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/{id} [get]
func (h *handler) FirmwareGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}

	firmware, err := h.firmwareUseCase.ReadById(c.Request().Context(), domainusecasesnode.ReadFirmwareByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the firmware. Please try again.")
	}
	if firmware == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("firmware"), "The requested firmware could not be found.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Firmware(*firmware))
}

// FirmwarePatch godoc
//
// @Summary Firmware
// @Tags Firmwares
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.FirmwarePatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/{id} [patch]
func (h *handler) FirmwarePatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}

	var req presentationhttprequest.FirmwarePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	nodeClassId, err := presentationhttputils.OptionalUUID(req.NodeClassId, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node class ID provided is invalid.")
	}

	if err := h.firmwareUseCase.UpdateById(c.Request().Context(), domainusecasesnode.UpdateFirmwareRequest{
		Id:          id,
		NodeClassId: nodeClassId,
		Name:        req.Name,
		UpdatedBy:   presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to update the firmware. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// FirmwareBinaryPut godoc
//
// @Summary Firmware Binary Put
// @Tags Firmwares
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param file formData file true "file"
// @Param config_schema formData string false "config_schema"
// @Success 200 {object} presentationhttpresponse.FirmwareBinaryStatResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/{id}/binary [put]
func (h *handler) FirmwareBinaryPut(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}
	file, err := presentationhttputils.RequiredFormFile(c, "file")
	if err != nil {
		return presentationhttputils.Error(c, err, "A firmware file is required.")
	}

	content, err := file.Open()
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to read the uploaded file. Please try again.")
	}
	defer content.Close()

	configSchema, err := parseConfigSchemaFormValue(c.FormValue("config_schema"))
	if err != nil {
		return presentationhttputils.Error(c, err, "The config_schema field must be a valid JSON array of {key,value_type} objects.")
	}

	stat, err := h.firmwareUseCase.ReplaceBinaryById(c.Request().Context(), domainusecasesnode.ReplaceFirmwareBinaryByIdRequest{
		Id:           id,
		Content:      content,
		ConfigSchema: configSchema,
		UpdatedBy:    presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to upload the firmware binary. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.FirmwareBinaryStatResponse{
		BinaryPath: stat.BinaryPath,
		Size:       stat.Size,
		Checksum:   stat.Checksum,
	})
}

// FirmwareBinaryGetById godoc
//
// @Summary Firmware Binary Get By ID
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 302 "redirect to a short-lived presigned download URL"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/{id}/binary [get]
func (h *handler) FirmwareBinaryGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}

	result, err := h.firmwareUseCase.DownloadUrlById(c.Request().Context(), domainusecasesnode.DownloadFirmwareBinaryByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to generate a download link for this firmware. Please try again.")
	}

	return c.Redirect(http.StatusFound, result.DownloadUrl)
}

// FirmwareBinaryGetByName godoc
//
// @Summary Firmware Binary Get By Name
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 302 "redirect to a short-lived presigned download URL"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/by-name/{name}/binary [get]
func (h *handler) FirmwareBinaryGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid firmware name.")
	}

	result, err := h.firmwareUseCase.DownloadUrlByName(c.Request().Context(), domainusecasesnode.DownloadFirmwareBinaryByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to generate a download link for this firmware. Please try again.")
	}

	return c.Redirect(http.StatusFound, result.DownloadUrl)
}

// FirmwareBinaryStatByNameGet godoc
//
// @Summary Firmware Binary Stat By Name Get
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200 {object} presentationhttpresponse.FirmwareBinaryStatResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/by-name/{name}/binary/stat [get]
func (h *handler) FirmwareBinaryStatByNameGet(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid firmware name.")
	}

	stat, err := h.firmwareUseCase.StatBinaryByName(c.Request().Context(), domainusecasesnode.StatFirmwareBinaryByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the firmware binary. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.FirmwareBinaryStatResponse{
		BinaryPath: stat.BinaryPath,
		Size:       stat.Size,
		Checksum:   stat.Checksum,
	})
}

// FirmwareBinaryStatByNameHead godoc
//
// @Summary Firmware Binary Stat By Name Head
// @Tags Firmwares
// @Produce octet-stream
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/by-name/{name}/binary [head]
func (h *handler) FirmwareBinaryStatByNameHead(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid firmware name.")
	}

	stat, err := h.firmwareUseCase.StatBinaryByName(c.Request().Context(), domainusecasesnode.StatFirmwareBinaryByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to look up the firmware binary. Please try again.")
	}

	headers := c.Response().Header()
	headers.Set("Content-Length", strconv.FormatInt(int64(stat.Size), 10))
	headers.Set("X-Firmware-Checksum", stat.Checksum)
	headers.Set("X-Firmware-Binary-Path", stat.BinaryPath)

	return c.NoContent(http.StatusOK)
}

// FirmwareDelete godoc
//
// @Summary Firmware Delete
// @Tags Firmwares
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.FirmwareDeleteRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/{id} [delete]
func (h *handler) FirmwareDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}

	var req presentationhttprequest.FirmwareDeleteRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	expectedName, err := presentationhttputils.RequiredString(req.ExpectedName, "expected_name")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware name confirmation is required.")
	}

	if err := h.firmwareUseCase.DeleteById(c.Request().Context(), domainusecasesnode.DeleteFirmwareRequest{
		Id:           id,
		ExpectedName: expectedName,
		DeletedBy:    presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete the firmware. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// OtaDispatchByNodeIdPost godoc
//
// @Summary Ota Dispatch By Node ID
// @Tags OTA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.OtaDispatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Failure 504 {object} presentationhttpresponse.ErrorResponse "Request Timeout"
// @Router /v1/nodes/{id}/ota [post]
func (h *handler) OtaDispatchByNodeIdPost(c *echo.Context) error {
	nodeId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}

	var req presentationhttprequest.OtaDispatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	firmwareId, err := otaRequest(req)
	if err != nil {
		return presentationhttputils.Error(c, err, "Please select a valid firmware.")
	}

	if err := h.otaUseCase.DispatchByNodeId(c.Request().Context(), domainusecasesnode.DispatchOtaByNodeIdRequest{
		NodeId:     nodeId,
		FirmwareId: firmwareId,
		ActorId:    presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to dispatch the firmware update. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

// OtaDispatchByNodeDeviceIdPost godoc
//
// @Summary Ota Dispatch By Node Device ID
// @Tags OTA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param device_id path string true "device_id"
// @Param request body presentationhttprequest.OtaDispatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Failure 504 {object} presentationhttpresponse.ErrorResponse "Request Timeout"
// @Router /v1/nodes/by-device/{device_id}/ota [post]
func (h *handler) OtaDispatchByNodeDeviceIdPost(c *echo.Context) error {
	deviceId, err := presentationhttputils.RequiredString(c.Param("device_id"), "device_id")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a valid device ID.")
	}

	var req presentationhttprequest.OtaDispatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	firmwareId, err := otaRequest(req)
	if err != nil {
		return presentationhttputils.Error(c, err, "Please select a valid firmware.")
	}

	if err := h.otaUseCase.DispatchByNodeDeviceId(c.Request().Context(), domainusecasesnode.DispatchOtaByNodeDeviceIdRequest{
		NodeDeviceId: deviceId,
		FirmwareId:   firmwareId,
		ActorId:      presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to dispatch the firmware update. Please try again.")
	}

	return c.NoContent(http.StatusNoContent)
}

func otaRequest(req presentationhttprequest.OtaDispatchRequest) (uuid.UUID, error) {
	firmwareId, err := presentationhttputils.RequiredUUID(req.FirmwareId, "firmware_id")
	if err != nil {
		return uuid.Nil, err
	}

	return firmwareId, nil
}

func parseConfigSchemaFormValue(raw string) ([]domainusecasesnode.ConfigParameterInput, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var items []presentationhttprequest.FirmwareConfigSchemaItemRequest
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, domainmodels.NewError("config_schema must be a valid JSON array", domainmodels.ErrTypeValidation, err)
	}

	schema := make([]domainusecasesnode.ConfigParameterInput, 0, len(items))
	for _, item := range items {
		schema = append(schema, domainusecasesnode.ConfigParameterInput{
			Key:       item.Key,
			ValueType: item.ValueType,
		})
	}

	return schema, nil
}

// FirmwareConfigParametersGet godoc
//
// @Summary Firmware Config Parameters
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.FirmwareConfigParameterResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/firmwares/{id}/config-parameters [get]
func (h *handler) FirmwareConfigParametersGet(c *echo.Context) error {
	firmwareId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The firmware ID provided is invalid.")
	}

	params, err := h.configParameterUseCase.ReadByFirmwareId(c.Request().Context(), domainusecasesnode.ReadConfigParametersByFirmwareIdRequest{
		FirmwareId: firmwareId,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to read the firmware's config parameters. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.FirmwareConfigParameters(params))
}

// NodeConfigGet godoc
//
// @Summary Node Config
// @Tags Nodes
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.NodeConfigValueResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/{id}/config [get]
func (h *handler) NodeConfigGet(c *echo.Context) error {
	nodeId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}

	values, err := h.configValueUseCase.ReadByNodeId(c.Request().Context(), domainusecasesnode.ReadConfigValuesByNodeIdRequest{
		NodeId: nodeId,
	})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to read the node's config values. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.NodeConfigValues(values))
}

// NodeConfigPut godoc
//
// @Summary Node Config
// @Tags Nodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.SetNodeConfigValueRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/nodes/{id}/config [put]
func (h *handler) NodeConfigPut(c *echo.Context) error {
	nodeId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err, "The node ID provided is invalid.")
	}

	var req presentationhttprequest.SetNodeConfigValueRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	key, err := presentationhttputils.RequiredString(req.Key, "key")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please provide a config key.")
	}
	if req.Value == nil {
		return presentationhttputils.Error(
			c,
			domainmodels.NewError("value is required", domainmodels.ErrTypeValidation, nil),
			"Please provide a config value.",
		)
	}

	if err := h.configValueUseCase.SetByNodeId(c.Request().Context(), domainusecasesnode.SetConfigValueRequest{
		NodeId:  nodeId,
		Key:     key,
		Value:   *req.Value,
		ActorId: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to set the node's config value. Please check your input and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}
