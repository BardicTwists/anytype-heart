package object

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/anyproto/anytype-heart/pb"
	"github.com/anyproto/anytype-heart/pb/service/mock_service"
	"github.com/anyproto/anytype-heart/pkg/lib/bundle"
	"github.com/anyproto/anytype-heart/pkg/lib/pb/model"
	"github.com/anyproto/anytype-heart/util/pbtypes"
)

const (
	offset            = 0
	limit             = 100
	mockedSpaceId     = "mocked-space-id"
	mockedObjectId    = "mocked-object-id"
	mockedNewObjectId = "mocked-new-object-id"
	mockedTechSpaceId = "mocked-tech-space-id"
	gatewayUrl        = "http://localhost:31006"
)

type fixture struct {
	*ObjectService
	mwMock *mock_service.MockClientCommandsServer
}

func newFixture(t *testing.T) *fixture {
	mw := mock_service.NewMockClientCommandsServer(t)
	objectService := NewService(mw)
	objectService.AccountInfo = &model.AccountInfo{
		TechSpaceId: mockedTechSpaceId,
		GatewayUrl:  gatewayUrl,
	}

	return &fixture{
		ObjectService: objectService,
		mwMock:        mw,
	}
}

func TestObjectService_ListObjects(t *testing.T) {
	t.Run("successfully get objects for a space", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectSearch", mock.Anything, &pb.RpcObjectSearchRequest{
			SpaceId: mockedSpaceId,
			Filters: []*model.BlockContentDataviewFilter{
				{
					RelationKey: bundle.RelationKeyLayout.String(),
					Condition:   model.BlockContentDataviewFilter_In,
					Value: pbtypes.IntList([]int{
						int(model.ObjectType_basic),
						int(model.ObjectType_profile),
						int(model.ObjectType_todo),
						int(model.ObjectType_note),
						int(model.ObjectType_bookmark),
						int(model.ObjectType_set),
						int(model.ObjectType_collection),
						int(model.ObjectType_participant),
					}...),
				},
			},
			Sorts: []*model.BlockContentDataviewSort{{
				RelationKey:    bundle.RelationKeyLastModifiedDate.String(),
				Type:           model.BlockContentDataviewSort_Desc,
				Format:         model.RelationFormat_longtext,
				IncludeTime:    true,
				EmptyPlacement: model.BlockContentDataviewSort_NotSpecified,
			}},
			FullText:         "",
			Offset:           0,
			Limit:            0,
			ObjectTypeFilter: []string{},
			Keys:             []string{"id", "name"},
		}).Return(&pb.RpcObjectSearchResponse{
			Records: []*types.Struct{
				{
					Fields: map[string]*types.Value{
						"id":        pbtypes.String(mockedObjectId),
						"name":      pbtypes.String("My Object"),
						"type":      pbtypes.String("ot-page"),
						"layout":    pbtypes.Float64(float64(model.ObjectType_basic)),
						"iconEmoji": pbtypes.String("📄"),
					},
				},
			},
			Error: &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
		}).Once()

		// Mock object show for object details
		fx.mwMock.On("ObjectShow", mock.Anything, &pb.RpcObjectShowRequest{
			SpaceId:  mockedSpaceId,
			ObjectId: mockedObjectId,
		}).Return(&pb.RpcObjectShowResponse{
			ObjectView: &model.ObjectView{
				RootId: mockedObjectId,
				Details: []*model.ObjectViewDetailsSet{
					{
						Details: &types.Struct{
							Fields: map[string]*types.Value{
								"id":               pbtypes.String(mockedObjectId),
								"name":             pbtypes.String("My Object"),
								"type":             pbtypes.String("ot-page"),
								"iconEmoji":        pbtypes.String("📄"),
								"lastModifiedDate": pbtypes.Float64(999999),
								"createdDate":      pbtypes.Float64(888888),
							},
						},
					},
				},
			},
			Error: &pb.RpcObjectShowResponseError{Code: pb.RpcObjectShowResponseError_NULL},
		}).Once()

		// Mock type resolution
		fx.mwMock.On("ObjectSearch", mock.Anything, &pb.RpcObjectSearchRequest{
			SpaceId: mockedSpaceId,
			Filters: []*model.BlockContentDataviewFilter{
				{
					RelationKey: "uniqueKey",
					Condition:   model.BlockContentDataviewFilter_Equal,
					Value:       pbtypes.String("ot-page"),
				},
			},
			Keys: []string{"name"},
		}).Return(&pb.RpcObjectSearchResponse{
			Records: []*types.Struct{
				{
					Fields: map[string]*types.Value{
						"name": pbtypes.String("Page"),
					},
				},
			},
			Error: &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
		}).Once()

		// when
		objects, total, hasMore, err := fx.ListObjects(ctx, mockedSpaceId, offset, limit)

		// then
		require.NoError(t, err)
		require.Len(t, objects, 1)
		require.Equal(t, mockedObjectId, objects[0].Id)
		require.Equal(t, "My Object", objects[0].Name)
		require.Equal(t, "Page", objects[0].ObjectType)
		require.Equal(t, "📄", objects[0].Icon)
		require.Equal(t, 3, len(objects[0].Details))

		for _, detail := range objects[0].Details {
			if detail.Id == "createdDate" {
				require.Equal(t, float64(888888), detail.Details["createdDate"])
			} else if detail.Id == "lastModifiedDate" {
				require.Equal(t, float64(999999), detail.Details["lastModifiedDate"])
			} else if detail.Id == "tags" {
				require.Empty(t, detail.Details["tags"])
			} else {
				t.Errorf("unexpected detail id: %s", detail.Id)
			}
		}

		require.Equal(t, 1, total)
		require.False(t, hasMore)
	})

	t.Run("no objects found", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectSearch", mock.Anything, mock.Anything).
			Return(&pb.RpcObjectSearchResponse{
				Records: []*types.Struct{},
				Error:   &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
			}).Once()

		// when
		objects, total, hasMore, err := fx.ListObjects(ctx, "empty-space", offset, limit)

		// then
		require.ErrorIs(t, err, ErrNoObjectsFound)
		require.Len(t, objects, 0)
		require.Equal(t, 0, total)
		require.False(t, hasMore)
	})
}

func TestObjectService_GetObject(t *testing.T) {
	t.Run("object found", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectShow", mock.Anything, &pb.RpcObjectShowRequest{
			SpaceId:  mockedSpaceId,
			ObjectId: mockedObjectId,
		}).
			Return(&pb.RpcObjectShowResponse{
				Error: &pb.RpcObjectShowResponseError{Code: pb.RpcObjectShowResponseError_NULL},
				ObjectView: &model.ObjectView{
					RootId: mockedObjectId,
					Details: []*model.ObjectViewDetailsSet{
						{
							Details: &types.Struct{
								Fields: map[string]*types.Value{
									"id":               pbtypes.String(mockedObjectId),
									"name":             pbtypes.String("Found Object"),
									"type":             pbtypes.String("ot-page"),
									"iconEmoji":        pbtypes.String("🔍"),
									"lastModifiedDate": pbtypes.Float64(999999),
									"createdDate":      pbtypes.Float64(888888),
								},
							},
						},
					},
				},
			}, nil).Once()

		// Mock type resolution
		fx.mwMock.On("ObjectSearch", mock.Anything, &pb.RpcObjectSearchRequest{
			SpaceId: mockedSpaceId,
			Filters: []*model.BlockContentDataviewFilter{
				{
					RelationKey: "uniqueKey",
					Condition:   model.BlockContentDataviewFilter_Equal,
					Value:       pbtypes.String("ot-page"),
				},
			},
			Keys: []string{"name"},
		}).Return(&pb.RpcObjectSearchResponse{
			Error: &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
			Records: []*types.Struct{
				{
					Fields: map[string]*types.Value{
						"name": pbtypes.String("Page"),
					},
				},
			},
		}, nil).Once()

		// when
		object, err := fx.GetObject(ctx, mockedSpaceId, mockedObjectId)

		// then
		require.NoError(t, err)
		require.Equal(t, mockedObjectId, object.Id)
		require.Equal(t, "Found Object", object.Name)
		require.Equal(t, "Page", object.ObjectType)
		require.Equal(t, "🔍", object.Icon)
		require.Equal(t, 3, len(object.Details))

		for _, detail := range object.Details {
			if detail.Id == "createdDate" {
				require.Equal(t, float64(888888), detail.Details["createdDate"])
			} else if detail.Id == "lastModifiedDate" {
				require.Equal(t, float64(999999), detail.Details["lastModifiedDate"])
			} else if detail.Id == "tags" {
				require.Empty(t, detail.Details["tags"])
			} else {
				t.Errorf("unexpected detail id: %s", detail.Id)
			}
		}
	})

	t.Run("object not found", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectShow", mock.Anything, mock.Anything).
			Return(&pb.RpcObjectShowResponse{
				Error: &pb.RpcObjectShowResponseError{Code: pb.RpcObjectShowResponseError_NOT_FOUND},
			}, nil).Once()

		// when
		object, err := fx.GetObject(ctx, mockedSpaceId, "missing-obj")

		// then
		require.ErrorIs(t, err, ErrObjectNotFound)
		require.Empty(t, object)
	})
}

func TestObjectService_CreateObject(t *testing.T) {
	t.Run("successful object creation", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectCreate", mock.Anything, &pb.RpcObjectCreateRequest{
			Details: &types.Struct{
				Fields: map[string]*types.Value{
					"name":      pbtypes.String("New Object"),
					"iconEmoji": pbtypes.String("🆕"),
				},
			},
			TemplateId:          "",
			SpaceId:             mockedSpaceId,
			ObjectTypeUniqueKey: "",
			WithChat:            false,
		}).Return(&pb.RpcObjectCreateResponse{
			ObjectId: mockedNewObjectId,
			Details: &types.Struct{
				Fields: map[string]*types.Value{
					"id":        pbtypes.String(mockedNewObjectId),
					"name":      pbtypes.String("New Object"),
					"iconEmoji": pbtypes.String("🆕"),
					"spaceId":   pbtypes.String(mockedSpaceId),
				},
			},
			Error: &pb.RpcObjectCreateResponseError{Code: pb.RpcObjectCreateResponseError_NULL},
		}).Once()

		// Mock object show for object details
		fx.mwMock.On("ObjectShow", mock.Anything, &pb.RpcObjectShowRequest{
			SpaceId:  mockedSpaceId,
			ObjectId: mockedNewObjectId,
		}).Return(&pb.RpcObjectShowResponse{
			ObjectView: &model.ObjectView{
				RootId: mockedNewObjectId,
				Details: []*model.ObjectViewDetailsSet{
					{
						Details: &types.Struct{
							Fields: map[string]*types.Value{
								"id":        pbtypes.String(mockedNewObjectId),
								"name":      pbtypes.String("New Object"),
								"type":      pbtypes.String("ot-page"),
								"iconEmoji": pbtypes.String("🆕"),
								"spaceId":   pbtypes.String(mockedSpaceId),
							},
						},
					},
				},
			},
			Error: &pb.RpcObjectShowResponseError{Code: pb.RpcObjectShowResponseError_NULL},
		}).Once()

		// Mock type resolution
		fx.mwMock.On("ObjectSearch", mock.Anything, &pb.RpcObjectSearchRequest{
			SpaceId: mockedSpaceId,
			Filters: []*model.BlockContentDataviewFilter{
				{
					RelationKey: "uniqueKey",
					Condition:   model.BlockContentDataviewFilter_Equal,
					Value:       pbtypes.String("ot-page"),
				},
			},
			Keys: []string{"name"},
		}).Return(&pb.RpcObjectSearchResponse{
			Error: &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
			Records: []*types.Struct{
				{
					Fields: map[string]*types.Value{
						"name": pbtypes.String("Page"),
					},
				},
			},
		}).Once()

		// when
		object, err := fx.CreateObject(ctx, mockedSpaceId, CreateObjectRequest{
			Name: "New Object",
			Icon: "🆕",
			// TODO: use actual values
			TemplateId:          "",
			ObjectTypeUniqueKey: "",
			WithChat:            false,
		})

		// then
		require.NoError(t, err)
		require.Equal(t, mockedNewObjectId, object.Id)
		require.Equal(t, "New Object", object.Name)
		require.Equal(t, "Page", object.ObjectType)
		require.Equal(t, "🆕", object.Icon)
		require.Equal(t, mockedSpaceId, object.SpaceId)
	})

	t.Run("creation error", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectCreate", mock.Anything, mock.Anything).
			Return(&pb.RpcObjectCreateResponse{
				Error: &pb.RpcObjectCreateResponseError{Code: pb.RpcObjectCreateResponseError_UNKNOWN_ERROR},
			}).Once()

		// when
		object, err := fx.CreateObject(ctx, mockedSpaceId, CreateObjectRequest{
			Name: "Fail Object",
			Icon: "",
		})

		// then
		require.ErrorIs(t, err, ErrFailedCreateObject)
		require.Empty(t, object)
	})
}

func TestObjectService_UpdateObject(t *testing.T) {
	t.Run("not implemented", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		// when
		object, err := fx.UpdateObject(ctx, mockedSpaceId, mockedObjectId, UpdateObjectRequest{
			Object: Object{
				Name: "Updated Object",
			},
		})

		// then
		require.ErrorIs(t, err, ErrNotImplemented)
		require.Empty(t, object)
	})

	// TODO: further tests
}

func TestObjectService_ListTypes(t *testing.T) {
	t.Run("types found", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectSearch", mock.Anything, mock.Anything).
			Return(&pb.RpcObjectSearchResponse{
				Records: []*types.Struct{
					{
						Fields: map[string]*types.Value{
							"id":        pbtypes.String("type-1"),
							"name":      pbtypes.String("Type One"),
							"uniqueKey": pbtypes.String("type-one-key"),
							"iconEmoji": pbtypes.String("🗂️"),
						},
					},
				},
				Error: &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
			}).Once()

		// when
		types, total, hasMore, err := fx.ListTypes(ctx, mockedSpaceId, offset, limit)

		// then
		require.NoError(t, err)
		require.Len(t, types, 1)
		require.Equal(t, "type-1", types[0].Id)
		require.Equal(t, "Type One", types[0].Name)
		require.Equal(t, "type-one-key", types[0].UniqueKey)
		require.Equal(t, "🗂️", types[0].Icon)
		require.Equal(t, 1, total)
		require.False(t, hasMore)
	})

	t.Run("no types found", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectSearch", mock.Anything, mock.Anything).
			Return(&pb.RpcObjectSearchResponse{
				Records: []*types.Struct{},
				Error:   &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
			}).Once()

		// when
		types, total, hasMore, err := fx.ListTypes(ctx, "empty-space", offset, limit)

		// then
		require.ErrorIs(t, err, ErrNoTypesFound)
		require.Len(t, types, 0)
		require.Equal(t, 0, total)
		require.False(t, hasMore)
	})
}

func TestObjectService_ListTemplates(t *testing.T) {
	t.Run("templates found", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		// Mock template type search
		fx.mwMock.On("ObjectSearch", mock.Anything, mock.Anything).Return(&pb.RpcObjectSearchResponse{
			Records: []*types.Struct{
				{
					Fields: map[string]*types.Value{
						"id":        pbtypes.String("template-type-id"),
						"uniqueKey": pbtypes.String("ot-template"),
					},
				},
			},
			Error: &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
		}).Once()

		// Mock actual template objects search
		fx.mwMock.On("ObjectSearch", mock.Anything, mock.Anything).Return(&pb.RpcObjectSearchResponse{
			Records: []*types.Struct{
				{
					Fields: map[string]*types.Value{
						"id":               pbtypes.String("template-1"),
						"targetObjectType": pbtypes.String("target-type-id"),
					},
				},
			},
			Error: &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
		}).Once()

		// Mock object show for template details
		fx.mwMock.On("ObjectShow", mock.Anything, mock.Anything).Return(&pb.RpcObjectShowResponse{
			Error: &pb.RpcObjectShowResponseError{Code: pb.RpcObjectShowResponseError_NULL},
			ObjectView: &model.ObjectView{
				Details: []*model.ObjectViewDetailsSet{
					{
						Details: &types.Struct{
							Fields: map[string]*types.Value{
								"name":      pbtypes.String("Template Name"),
								"iconEmoji": pbtypes.String("📝"),
							},
						},
					},
				},
			},
		}, nil).Once()

		// when
		templates, total, hasMore, err := fx.ListTemplates(ctx, mockedSpaceId, "target-type-id", offset, limit)

		// then
		require.NoError(t, err)
		require.Len(t, templates, 1)
		require.Equal(t, "template-1", templates[0].Id)
		require.Equal(t, "Template Name", templates[0].Name)
		require.Equal(t, "📝", templates[0].Icon)
		require.Equal(t, 1, total)
		require.False(t, hasMore)
	})

	t.Run("no template type found", func(t *testing.T) {
		// given
		ctx := context.Background()
		fx := newFixture(t)

		fx.mwMock.On("ObjectSearch", mock.Anything, mock.Anything).
			Return(&pb.RpcObjectSearchResponse{
				Records: []*types.Struct{},
				Error:   &pb.RpcObjectSearchResponseError{Code: pb.RpcObjectSearchResponseError_NULL},
			}).Once()

		// when
		templates, total, hasMore, err := fx.ListTemplates(ctx, mockedSpaceId, "missing-type-id", offset, limit)

		// then
		require.ErrorIs(t, err, ErrTemplateTypeNotFound)
		require.Len(t, templates, 0)
		require.Equal(t, 0, total)
		require.False(t, hasMore)
	})
}
