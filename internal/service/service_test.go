package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/devWaylander/pvz_store/api"
	internalErrors "github.com/devWaylander/pvz_store/pkg/errors"
	"github.com/devWaylander/pvz_store/pkg/models"
	"github.com/google/uuid"
)

func Test_service_CreatePVZ(t *testing.T) {
	newUuid := uuid.New()
	newTime := time.Now()

	type fields struct {
		repo Repository
	}
	type args struct {
		ctx  context.Context
		data api.PVZ
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    api.PVZ
		wantErr bool
	}{
		{
			name: "Create new PVZ",
			fields: fields{
				repo: &MockRepository{
					CreatePVZFunc: func(ctx context.Context, id uuid.UUID, city string, registrationDate time.Time) (api.PVZ, error) {
						return api.PVZ{
							Id:               &id,
							City:             api.PVZCity(city),
							RegistrationDate: &registrationDate,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				data: api.PVZ{
					Id:               &newUuid,
					City:             "Test City",
					RegistrationDate: &newTime,
				},
			},
			want: api.PVZ{
				Id:               &newUuid,
				City:             "Test City",
				RegistrationDate: &newTime,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo: tt.fields.repo,
			}
			got, err := s.CreatePVZ(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreatePVZ() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.CreatePVZ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_GetPVZsInfo(t *testing.T) {
	page := 1
	limit := 10
	newUuid := uuid.New()
	newTime := time.Now()

	type fields struct {
		repo Repository
	}
	type args struct {
		ctx  context.Context
		data api.GetPvzParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.PvzInfo
		wantErr bool
	}{
		{
			name: "Get PVZ info",
			fields: fields{
				repo: &MockRepository{
					// Мок для получения списка PVZ
					GetPVZsFunc: func(ctx context.Context, page, limit int) ([]api.PVZ, error) {
						return []api.PVZ{
							{
								Id:               &newUuid,
								City:             "Test City",
								RegistrationDate: &newTime,
							},
						}, nil
					},
					// Мок для получения приемок по UUID PVZ
					GetReceptionsByPvzUUIDsFilteredFunc: func(ctx context.Context, pvzUUIDs []uuid.UUID, startDate, endDate *time.Time) ([]api.Reception, error) {
						return []api.Reception{
							{
								Id:     &newUuid,
								PvzId:  newUuid,
								Status: api.InProgress,
							},
						}, nil
					},
					// Мок для получения продуктов по UUID приемок
					GetProductsByRecsUUIDsFunc: func(ctx context.Context, recsUUIDs []uuid.UUID) ([]api.Product, error) {
						return []api.Product{
							{
								Id:          &newUuid,
								ReceptionId: newUuid,
								Type:        "ProductType",
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				data: api.GetPvzParams{
					Page:  &page,
					Limit: &limit,
				},
			},
			want: []models.PvzInfo{
				{
					Pvz: api.PVZ{
						Id:               &newUuid,
						City:             "Test City",
						RegistrationDate: &newTime,
					},
					Receptions: []models.ReceptionWithProducts{
						{
							Reception: api.Reception{
								Id:     &newUuid,
								PvzId:  newUuid,
								Status: api.InProgress,
							},
							Products: []api.Product{
								{
									Id:          &newUuid,
									ReceptionId: newUuid,
									Type:        "ProductType",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo: tt.fields.repo,
			}
			got, err := s.GetPVZsInfo(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetPVZsInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.GetPVZsInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_CreateReception(t *testing.T) {
	newUuid := uuid.New()

	type fields struct {
		repo Repository
	}
	type args struct {
		ctx  context.Context
		data api.PostReceptionsJSONBody
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    api.Reception
		wantErr bool
	}{
		{
			name: "Create Reception",
			fields: fields{
				repo: &MockRepository{
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						// Имитация того, что PVZ существует
						return true, nil
					},
					GetReceptionStatusByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (string, error) {
						// Мок статуса приема
						return "opened", nil
					},
					CreateReceptionFunc: func(ctx context.Context, pvzUUID uuid.UUID, status string) (api.Reception, error) {
						// Возвращаем ожидаемое значение
						return api.Reception{
							Id:     &newUuid,
							PvzId:  pvzUUID,
							Status: "opened",
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				data: api.PostReceptionsJSONBody{
					PvzId: newUuid,
				},
			},
			want: api.Reception{
				Id:     &newUuid,
				PvzId:  newUuid,
				Status: "opened",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo: tt.fields.repo,
			}
			got, err := s.CreateReception(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreateReception() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.CreateReception() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_CloseReception(t *testing.T) {
	newUuid := uuid.New()

	type fields struct {
		repo Repository
	}
	type args struct {
		ctx     context.Context
		pvzUUID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    api.Reception
		wantErr bool
	}{
		{
			name: "Close Reception",
			fields: fields{
				repo: &MockRepository{
					// Возвращаем приёмку с открытым статусом
					GetReceptionByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
						return api.Reception{
							Id:     &newUuid,
							PvzId:  newUuid,
							Status: api.InProgress,
						}, nil
					},
					// Успешно обновляем статус на "closed"
					UpdateReceptionStatusFunc: func(ctx context.Context, recUUID uuid.UUID, status string) error {
						if status == string(api.Close) {
							return nil
						}
						return errors.New("status update failed")
					},
					// Добавляем mock для IsPVZExistFunc
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return true, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				pvzUUID: newUuid,
			},
			want: api.Reception{
				Id:     &newUuid,
				PvzId:  newUuid,
				Status: api.Close,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo: tt.fields.repo,
			}
			got, err := s.CloseReception(tt.args.ctx, tt.args.pvzUUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CloseReception() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.CloseReception() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_CreateProduct(t *testing.T) {
	newUuid := uuid.New()

	type fields struct {
		repo Repository
	}
	type args struct {
		ctx  context.Context
		data api.PostProductsJSONBody
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    api.Product
		wantErr bool
	}{
		{
			name: "Create Product",
			fields: fields{
				repo: &MockRepository{
					// Мок для проверки существования PVZ
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return true, nil // Возвращаем, что PVZ существует
					},
					// Мок для получения приёмки по UUID PVZ
					GetReceptionByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
						return api.Reception{
							Id:     &newUuid,
							PvzId:  newUuid,
							Status: api.InProgress, // Ожидаемый статус
						}, nil
					},
					// Мок для создания продукта
					CreateProductFunc: func(ctx context.Context, receptionUUID uuid.UUID, prType string) (api.Product, error) {
						return api.Product{
							Id: &newUuid, // Продукт с новым UUID
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				data: api.PostProductsJSONBody{
					PvzId: newUuid,
					Type:  api.PostProductsJSONBodyType("product"),
				},
			},
			want: api.Product{
				Id: &newUuid, // Ожидаем, что продукт будет создан с этим uuid
			},
			wantErr: false,
		},
		{
			name: "Error Creating Product - Repository Error",
			fields: fields{
				repo: &MockRepository{
					// Мокируем ошибку при создании продукта
					CreateProductFunc: func(ctx context.Context, receptionUUID uuid.UUID, prType string) (api.Product, error) {
						return api.Product{}, errors.New("create product failed")
					},
					// Мок для проверки существования PVZ
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return true, nil // Возвращаем, что PVZ существует
					},
					// Мок для получения приёмки по UUID PVZ
					GetReceptionByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
						return api.Reception{
							Id:     &newUuid,
							PvzId:  newUuid,
							Status: api.InProgress, // Ожидаемый статус
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				data: api.PostProductsJSONBody{
					PvzId: newUuid,
					Type:  api.PostProductsJSONBodyType("product"),
				},
			},
			want:    api.Product{}, // Ожидаем пустой продукт
			wantErr: true,          // Ошибка должна быть
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo: tt.fields.repo,
			}
			got, err := s.CreateProduct(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.CreateProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.CreateProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_getReceptionByPvzUUID(t *testing.T) {
	newUuid := uuid.New()
	type fields struct {
		repo Repository
	}
	type args struct {
		ctx     context.Context
		pvzUUID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    api.Reception
		wantErr bool
	}{
		{
			name: "Get Reception by PVZ UUID",
			fields: fields{
				repo: &MockRepository{
					// Мок для проверки существования PVZ
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return true, nil // Возвращаем, что PVZ существует
					},
					// Мок для получения приёмки по UUID PVZ
					GetReceptionByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
						return api.Reception{
							Id:     &newUuid,
							PvzId:  newUuid,
							Status: api.InProgress, // Ожидаемый статус
						}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				pvzUUID: newUuid,
			},
			want: api.Reception{
				Id:     &newUuid,
				PvzId:  newUuid,
				Status: api.InProgress,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo: tt.fields.repo,
			}
			got, err := s.getReceptionByPvzUUID(tt.args.ctx, tt.args.pvzUUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.getReceptionByPvzUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.getReceptionByPvzUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_service_DeleteLastProduct(t *testing.T) {
	newUuid := uuid.New()

	type fields struct {
		repo Repository
	}
	type args struct {
		ctx     context.Context
		pvzUUID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Successful Deletion of Last Product",
			fields: fields{
				repo: &MockRepository{
					// Мок для проверки существования PVZ
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return true, nil
					},
					// Мок для получения приемки по UUID PVZ
					GetReceptionByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
						return api.Reception{
							Id:     &newUuid,
							PvzId:  newUuid,
							Status: api.InProgress, // Статус приемки в процессе
						}, nil
					},
					// Мок для удаления последнего продукта
					DeleteLastProductByReceptionUUIDFunc: func(ctx context.Context, receptionUUID uuid.UUID) error {
						return nil // Успешное удаление
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				pvzUUID: newUuid,
			},
			wantErr: false,
		},
		{
			name: "PVZ Does Not Exist",
			fields: fields{
				repo: &MockRepository{
					// Мок для проверки существования PVZ
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return false, nil // PVZ не существует
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				pvzUUID: newUuid,
			},
			wantErr: true, // Ожидаем ошибку, так как PVZ не существует
		},
		{
			name: "Reception Does Not Exist",
			fields: fields{
				repo: &MockRepository{
					// Мок для проверки существования PVZ
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return true, nil
					},
					// Мок для получения приемки по UUID PVZ
					GetReceptionByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
						return api.Reception{}, errors.New(internalErrors.ErrReceptionDoesntExist)
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				pvzUUID: newUuid,
			},
			wantErr: true, // Ожидаем ошибку, так как приемка не существует
		},
		{
			name: "Wrong Reception Status",
			fields: fields{
				repo: &MockRepository{
					// Мок для проверки существования PVZ
					IsPVZExistFunc: func(ctx context.Context, id uuid.UUID) (bool, error) {
						return true, nil
					},
					// Мок для получения приемки по UUID PVZ
					GetReceptionByPvzUUIDFunc: func(ctx context.Context, pvzUUID uuid.UUID) (api.Reception, error) {
						return api.Reception{
							Id:     &newUuid,
							PvzId:  newUuid,
							Status: "Completed", // Неверный статус
						}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				pvzUUID: newUuid,
			},
			wantErr: true, // Ожидаем ошибку, так как статус приемки "Completed"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &service{
				repo: tt.fields.repo,
			}
			if err := s.DeleteLastProduct(tt.args.ctx, tt.args.pvzUUID); (err != nil) != tt.wantErr {
				t.Errorf("service.DeleteLastProduct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
