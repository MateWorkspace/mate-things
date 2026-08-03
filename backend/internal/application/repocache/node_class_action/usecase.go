package applicationrepocachenodeclassaction

import (
	"context"

	domaincontractscache "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/cache"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesrepocache "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/repocache"
	"github.com/google/uuid"
)

type usecase struct {
	repository     domaincontractsrepository.NodeClassAction
	cache          domaincontractscache.NodeClassAction
	nodeClassCache domaincontractscache.NodeClass
	actionCache    domaincontractscache.Action
}

func NewRepoCacheImpl(
	repository domaincontractsrepository.NodeClassAction,
	cache domaincontractscache.NodeClassAction,
	nodeClassCache domaincontractscache.NodeClass,
	actionCache domaincontractscache.Action,
) domainusecasesrepocache.NodeClassAction {
	return &usecase{
		repository:     repository,
		cache:          cache,
		nodeClassCache: nodeClassCache,
		actionCache:    actionCache,
	}
}

func (u *usecase) Create(
	ctx context.Context,
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
	createdBy *uuid.UUID,
) (uuid.UUID, error) {
	id, err := u.repository.Create(ctx, nodeClassId, actionId, createdBy)
	if err != nil {
		return uuid.Nil, err
	}

	if err := u.invalidateRelations(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (u *usecase) ReadById(
	ctx context.Context,
	id uuid.UUID,
) (*domainmodels.NodeClassAction, *domainmodels.NodeClass, *domainmodels.Action, error) {
	if item, hit, err := u.cache.GetById(ctx, id); err != nil {
		return nil, nil, nil, err
	} else if hit {
		return &item.NodeClassAction, &item.NodeClass, &item.Action, nil
	}

	nodeClassAction, nodeClass, action, err := u.repository.ReadById(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := u.cache.SetById(ctx, id, &domaincontractscache.NodeClassActionItem{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}); err != nil {
		return nil, nil, nil, err
	}

	return nodeClassAction, nodeClass, action, nil
}

func (u *usecase) ReadByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId uuid.UUID,
	actionId uuid.UUID,
) (*domainmodels.NodeClassAction, *domainmodels.NodeClass, *domainmodels.Action, error) {
	if item, hit, err := u.cache.GetByNodeClassIdAndActionId(ctx, nodeClassId, actionId); err != nil {
		return nil, nil, nil, err
	} else if hit {
		return &item.NodeClassAction, &item.NodeClass, &item.Action, nil
	}

	nodeClassAction, nodeClass, action, err := u.repository.ReadByNodeClassIdAndActionId(ctx, nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := u.cache.SetByNodeClassIdAndActionId(ctx, nodeClassId, actionId, &domaincontractscache.NodeClassActionItem{
		NodeClassAction: *nodeClassAction,
		NodeClass:       *nodeClass,
		Action:          *action,
	}); err != nil {
		return nil, nil, nil, err
	}

	return nodeClassAction, nodeClass, action, nil
}

func (u *usecase) ReadByPagination(
	ctx context.Context,
	page int,
	limit int,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action, int, error) {
	if pagination, hit, err := u.cache.GetPagination(ctx, page, limit, nodeClassId, actionId); err != nil {
		return nil, nil, nil, 0, err
	} else if hit {
		nodeClassActions, nodeClasses, actions := splitNodeClassActionItems(pagination.Items)
		return nodeClassActions, nodeClasses, actions, pagination.Total, nil
	}

	nodeClassActions, nodeClasses, actions, total, err := u.repository.ReadByPagination(ctx, page, limit, nodeClassId, actionId)
	if err != nil {
		return nil, nil, nil, 0, err
	}

	if err := u.cache.SetPagination(ctx, page, limit, nodeClassId, actionId, domaincontractscache.Pagination[domaincontractscache.NodeClassActionItem]{
		Items: joinNodeClassActionItems(nodeClassActions, nodeClasses, actions),
		Total: total,
	}); err != nil {
		return nil, nil, nil, 0, err
	}

	return nodeClassActions, nodeClasses, actions, total, nil
}

func (u *usecase) DeleteById(ctx context.Context, id uuid.UUID) error {
	if err := u.repository.DeleteById(ctx, id); err != nil {
		return err
	}

	return u.invalidateRelations(ctx)
}

func (u *usecase) DeleteByNodeClassIdAndActionId(
	ctx context.Context,
	nodeClassId *uuid.UUID,
	actionId *uuid.UUID,
) error {
	if err := u.repository.DeleteByNodeClassIdAndActionId(ctx, nodeClassId, actionId); err != nil {
		return err
	}

	return u.invalidateRelations(ctx)
}

// invalidateRelations clears this pivot's own cache, the owning node
// class's cached action list (NodeClass.ReadActions), and the actions
// list pagination cache (compatible_node_class_count is embedded there).
// Narrower than role_permission's fan-out: there's no downstream
// "computed grant" cache analog here.
func (u *usecase) invalidateRelations(ctx context.Context) error {
	if err := u.cache.InvalidateAll(ctx); err != nil {
		return err
	}
	if err := u.nodeClassCache.InvalidateActions(ctx); err != nil {
		return err
	}
	if err := u.actionCache.InvalidatePagination(ctx); err != nil {
		return err
	}
	return nil
}

func joinNodeClassActionItems(
	nodeClassActions []domainmodels.NodeClassAction,
	nodeClasses []domainmodels.NodeClass,
	actions []domainmodels.Action,
) []domaincontractscache.NodeClassActionItem {
	items := make([]domaincontractscache.NodeClassActionItem, len(nodeClassActions))
	for i := range nodeClassActions {
		items[i].NodeClassAction = nodeClassActions[i]
		if i < len(nodeClasses) {
			items[i].NodeClass = nodeClasses[i]
		}
		if i < len(actions) {
			items[i].Action = actions[i]
		}
	}
	return items
}

func splitNodeClassActionItems(
	items []domaincontractscache.NodeClassActionItem,
) ([]domainmodels.NodeClassAction, []domainmodels.NodeClass, []domainmodels.Action) {
	nodeClassActions := make([]domainmodels.NodeClassAction, len(items))
	nodeClasses := make([]domainmodels.NodeClass, len(items))
	actions := make([]domainmodels.Action, len(items))

	for i, item := range items {
		nodeClassActions[i] = item.NodeClassAction
		nodeClasses[i] = item.NodeClass
		actions[i] = item.Action
	}

	return nodeClassActions, nodeClasses, actions
}
