package presentationhttphandlernode

import (
	"mime"
	"net/http"
	"strconv"

	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	classUseCase    domainusecasesnode.ClassManagement
	deviceUseCase   domainusecasesnode.DeviceManagement
	firmwareUseCase domainusecasesnode.FirmwareManagement
	otaUseCase      domainusecasesnode.Ota
}

func NewHandler(
	classUseCase domainusecasesnode.ClassManagement,
	deviceUseCase domainusecasesnode.DeviceManagement,
	firmwareUseCase domainusecasesnode.FirmwareManagement,
	otaUseCase domainusecasesnode.Ota,
) *handler {
	return &handler{
		classUseCase:    classUseCase,
		deviceUseCase:   deviceUseCase,
		firmwareUseCase: firmwareUseCase,
		otaUseCase:      otaUseCase,
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
// @Success 201
// @Router /v1/node-classes [post]
func (h *handler) NodeClassPost(c *echo.Context) error {
	var req presentationhttprequest.NodeClassPostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	name, err := presentationhttputils.RequiredString(req.Name, "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	id, err := h.classUseCase.Create(c.Request().Context(), domainusecasesnode.CreateNodeClassRequest{
		Name:        name,
		Description: req.Description,
		CreatedBy:   presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// NodeClassGetList godoc
//
// @Summary Node Class List
// @Tags Node Classes
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /v1/node-classes [get]
func (h *handler) NodeClassGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	nodeClasses, total, err := h.classUseCase.ReadByPagination(c.Request().Context(), domainusecasesnode.ReadNodeClassesByPaginationRequest{
		Page:   page.Page,
		Limit:  page.Limit,
		Search: page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/node-classes/by-name/{name} [get]
func (h *handler) NodeClassGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	nodeClass, err := h.classUseCase.ReadByName(c.Request().Context(), domainusecasesnode.ReadNodeClassByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if nodeClass == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node class"))
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
// @Success 200
// @Router /v1/node-classes/{id} [get]
func (h *handler) NodeClassGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	nodeClass, err := h.classUseCase.ReadById(c.Request().Context(), domainusecasesnode.ReadNodeClassByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if nodeClass == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node class"))
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
// @Router /v1/node-classes/{id} [patch]
func (h *handler) NodeClassPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
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
		return presentationhttputils.Error(c, err)
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
// @Router /v1/node-classes/{id} [delete]
func (h *handler) NodeClassDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.classUseCase.DeleteById(c.Request().Context(), domainusecasesnode.DeleteNodeClassRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// NodeGetList godoc
//
// @Summary Node List
// @Tags Nodes
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /v1/nodes [get]
func (h *handler) NodeGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	nodeClassId, err := presentationhttputils.QueryUUID(c, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	nodes, total, err := h.deviceUseCase.ReadByPagination(c.Request().Context(), domainusecasesnode.ReadNodesByPaginationRequest{
		Page:         page.Page,
		Limit:        page.Limit,
		Search:       page.Search,
		NodeClassId:  nodeClassId,
		FirmwareName: presentationhttputils.QueryString(c, "firmware_name"),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/nodes/by-device/{device_id} [get]
func (h *handler) NodeGetByDeviceId(c *echo.Context) error {
	deviceId, err := presentationhttputils.RequiredString(c.Param("device_id"), "device_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	node, err := h.deviceUseCase.ReadByDeviceId(c.Request().Context(), domainusecasesnode.ReadNodeByDeviceIdRequest{DeviceId: deviceId})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if node == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node"))
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
// @Success 200
// @Router /v1/nodes/{id} [get]
func (h *handler) NodeGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	node, err := h.deviceUseCase.ReadById(c.Request().Context(), domainusecasesnode.ReadNodeByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if node == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("node"))
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
// @Router /v1/nodes/{id} [patch]
func (h *handler) NodePatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.NodePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	nodeClassId, err := presentationhttputils.OptionalUUID(req.NodeClassId, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.deviceUseCase.UpdateById(c.Request().Context(), domainusecasesnode.UpdateNodeRequest{
		Id:           id,
		NodeClassId:  nodeClassId,
		DeviceId:     req.DeviceId,
		Name:         req.Name,
		FirmwareName: req.FirmwareName,
		Description:  req.Description,
		UpdatedBy:    presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
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
// @Router /v1/nodes/{id}/firmware [patch]
func (h *handler) NodeFirmwarePatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.NodeFirmwarePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	firmwareName, err := presentationhttputils.RequiredString(req.FirmwareName, "firmware_name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.deviceUseCase.AssignFirmware(c.Request().Context(), domainusecasesnode.AssignNodeFirmwareRequest{
		Id:           id,
		FirmwareName: firmwareName,
		UpdatedBy:    presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
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
// @Router /v1/nodes/{id} [delete]
func (h *handler) NodeDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.deviceUseCase.DeleteById(c.Request().Context(), domainusecasesnode.DeleteNodeRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 201
// @Router /v1/firmwares [post]
func (h *handler) FirmwarePost(c *echo.Context) error {
	nodeClassId, err := presentationhttputils.RequiredUUID(c.FormValue("node_class_id"), "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	name, err := presentationhttputils.RequiredString(c.FormValue("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	file, err := presentationhttputils.RequiredFormFile(c, "file")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	content, err := file.Open()
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	defer content.Close()

	result, err := h.firmwareUseCase.Create(c.Request().Context(), domainusecasesnode.CreateFirmwareRequest{
		NodeClassId: nodeClassId,
		Name:        name,
		Content:     content,
		CreatedBy:   presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/firmwares [get]
func (h *handler) FirmwareGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	nodeClassId, err := presentationhttputils.QueryUUID(c, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	firmwares, total, err := h.firmwareUseCase.ReadByPagination(c.Request().Context(), domainusecasesnode.ReadFirmwaresByPaginationRequest{
		Page:        page.Page,
		Limit:       page.Limit,
		Search:      page.Search,
		NodeClassId: nodeClassId,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/node-classes/{node_class_id}/firmwares [get]
func (h *handler) FirmwareGetByNodeClassId(c *echo.Context) error {
	nodeClassId, err := presentationhttputils.RequiredUUID(c.Param("node_class_id"), "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	firmwares, total, err := h.firmwareUseCase.ReadByNodeClassIdAndPagination(c.Request().Context(), domainusecasesnode.ReadFirmwaresByNodeClassIdAndPaginationRequest{
		NodeClassId: nodeClassId,
		Page:        page.Page,
		Limit:       page.Limit,
		Search:      page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/nodes/{node_id}/firmwares/available [get]
func (h *handler) FirmwareGetAvailableByNodeId(c *echo.Context) error {
	nodeId, err := presentationhttputils.RequiredUUID(c.Param("node_id"), "node_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	firmwares, total, err := h.firmwareUseCase.ReadAvailableByNodeId(c.Request().Context(), domainusecasesnode.ReadAvailableFirmwaresByNodeIdRequest{
		NodeId: nodeId,
		Page:   page.Page,
		Limit:  page.Limit,
		Search: page.Search,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/firmwares/by-name/{name} [get]
func (h *handler) FirmwareGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	firmware, err := h.firmwareUseCase.ReadByName(c.Request().Context(), domainusecasesnode.ReadFirmwareByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if firmware == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("firmware"))
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
// @Success 200
// @Router /v1/firmwares/{id} [get]
func (h *handler) FirmwareGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	firmware, err := h.firmwareUseCase.ReadById(c.Request().Context(), domainusecasesnode.ReadFirmwareByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if firmware == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("firmware"))
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
// @Router /v1/firmwares/{id} [patch]
func (h *handler) FirmwarePatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.FirmwarePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	nodeClassId, err := presentationhttputils.OptionalUUID(req.NodeClassId, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.firmwareUseCase.UpdateById(c.Request().Context(), domainusecasesnode.UpdateFirmwareRequest{
		Id:          id,
		NodeClassId: nodeClassId,
		Name:        req.Name,
		UpdatedBy:   presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/firmwares/{id}/binary [put]
func (h *handler) FirmwareBinaryPut(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	file, err := presentationhttputils.RequiredFormFile(c, "file")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	content, err := file.Open()
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	defer content.Close()

	stat, err := h.firmwareUseCase.ReplaceBinaryById(c.Request().Context(), domainusecasesnode.ReplaceFirmwareBinaryByIdRequest{
		Id:        id,
		Content:   content,
		UpdatedBy: presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Produce octet-stream
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {file} file
// @Router /v1/firmwares/{id}/binary [get]
func (h *handler) FirmwareBinaryGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	result, err := h.firmwareUseCase.OpenBinaryById(c.Request().Context(), domainusecasesnode.OpenFirmwareBinaryByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return streamFirmware(c, result)
}

// FirmwareBinaryGetByName godoc
//
// @Summary Firmware Binary Get By Name
// @Tags Firmwares
// @Produce octet-stream
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200 {file} file
// @Router /v1/firmwares/by-name/{name}/binary [get]
func (h *handler) FirmwareBinaryGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	result, err := h.firmwareUseCase.OpenBinaryByName(c.Request().Context(), domainusecasesnode.OpenFirmwareBinaryByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return streamFirmware(c, result)
}

// FirmwareBinaryStatByNameGet godoc
//
// @Summary Firmware Binary Stat By Name Get
// @Tags Firmwares
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200
// @Router /v1/firmwares/by-name/{name}/binary/stat [get]
func (h *handler) FirmwareBinaryStatByNameGet(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	stat, err := h.firmwareUseCase.StatBinaryByName(c.Request().Context(), domainusecasesnode.StatFirmwareBinaryByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Router /v1/firmwares/by-name/{name}/binary [head]
func (h *handler) FirmwareBinaryStatByNameHead(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	stat, err := h.firmwareUseCase.StatBinaryByName(c.Request().Context(), domainusecasesnode.StatFirmwareBinaryByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err)
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
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Router /v1/firmwares/{id} [delete]
func (h *handler) FirmwareDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.firmwareUseCase.DeleteById(c.Request().Context(), domainusecasesnode.DeleteFirmwareRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/nodes/{id}/ota [post]
func (h *handler) OtaDispatchByNodeIdPost(c *echo.Context) error {
	nodeId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.OtaDispatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	firmwareName, firmwareUrl, err := otaRequest(req)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.otaUseCase.DispatchByNodeId(c.Request().Context(), domainusecasesnode.DispatchOtaByNodeIdRequest{
		NodeId:       nodeId,
		FirmwareName: firmwareName,
		FirmwareUrl:  firmwareUrl,
		ActorId:      presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
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
// @Success 200
// @Router /v1/nodes/by-device/{device_id}/ota [post]
func (h *handler) OtaDispatchByNodeDeviceIdPost(c *echo.Context) error {
	deviceId, err := presentationhttputils.RequiredString(c.Param("device_id"), "device_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.OtaDispatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	firmwareName, firmwareUrl, err := otaRequest(req)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.otaUseCase.DispatchByNodeDeviceId(c.Request().Context(), domainusecasesnode.DispatchOtaByNodeDeviceIdRequest{
		NodeDeviceId: deviceId,
		FirmwareName: firmwareName,
		FirmwareUrl:  firmwareUrl,
		ActorId:      presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func otaRequest(req presentationhttprequest.OtaDispatchRequest) (string, string, error) {
	firmwareName, err := presentationhttputils.RequiredString(req.FirmwareName, "firmware_name")
	if err != nil {
		return "", "", err
	}
	firmwareUrl, err := presentationhttputils.RequiredString(req.FirmwareUrl, "firmware_url")
	if err != nil {
		return "", "", err
	}

	return firmwareName, firmwareUrl, nil
}

func streamFirmware(c *echo.Context, result domainusecasesnode.OpenFirmwareBinaryResult) error {
	if result.Content == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("firmware content"))
	}
	defer result.Content.Close()

	filename := result.Firmware.Name
	if filename == "" {
		filename = result.Firmware.Id.String()
	}
	contentDisposition := mime.FormatMediaType("attachment", map[string]string{"filename": filename})
	headers := c.Response().Header()
	headers.Set("Content-Length", strconv.FormatInt(int64(result.Firmware.Size), 10))
	headers.Set("Content-Disposition", contentDisposition)
	headers.Set("X-Firmware-Checksum", result.Firmware.Checksum)
	headers.Set("X-Firmware-Name", result.Firmware.Name)

	return c.Stream(http.StatusOK, "application/octet-stream", result.Content)
}
