package main

import (
	"fireynis/velocity_checker/pkg/models"
	"fireynis/velocity_checker/pkg/models/mock"
	"fireynis/velocity_checker/pkg/validators"
	"testing"
	"time"
)

func Test_application_withinLimits(t *testing.T) {
	type fields struct {
		loads         models.ILoads
		loadValidator validators.ILoadValidator
	}
	type args struct {
		load *models.Load
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantErr      bool
		wantAccepted bool
	}{
		{
			name: "More than three loads daily",
			fields: fields{
				loads:         &mock.Load{},
				loadValidator: &validators.LoadValidator{},
			},
			args: args{
				load: &models.Load{
					Id:            0,
					TransactionId: 4,
					CustomerId:    1,
					Amount:        250000,
					Time:          time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
					Accepted:      false,
				},
			},
			wantErr:      false,
			wantAccepted: false,
		},
		{
			name: "More than 5k loaded daily",
			fields: fields{
				loads:         &mock.Load{},
				loadValidator: &validators.LoadValidator{},
			},
			args: args{
				load: &models.Load{
					Id:            0,
					TransactionId: 2,
					CustomerId:    2,
					Amount:        250000,
					Time:          time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
					Accepted:      false,
				},
			},
			wantErr:      false,
			wantAccepted: false,
		},
		{
			name: "More than 20k loaded weekly",
			fields: fields{
				loads:         &mock.Load{},
				loadValidator: &validators.LoadValidator{},
			},
			args: args{
				load: &models.Load{
					Id:            0,
					TransactionId: 2,
					CustomerId:    3,
					Amount:        250000,
					Time:          time.Date(2000, 1, 2, 16, 0, 0, 0, time.UTC),
					Accepted:      false,
				},
			},
			wantErr:      false,
			wantAccepted: false,
		},
		{
			name: "Acceptable load",
			fields: fields{
				loads:         &mock.Load{},
				loadValidator: &validators.LoadValidator{},
			},
			args: args{
				load: &models.Load{
					Id:            0,
					TransactionId: 2,
					CustomerId:    4,
					Amount:        250000,
					Time:          time.Date(2000, 1, 1, 16, 0, 0, 0, time.UTC),
					Accepted:      false,
				},
			},
			wantErr:      false,
			wantAccepted: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &application{
				loads:         tt.fields.loads,
				loadValidator: tt.fields.loadValidator,
			}
			if err := a.withinLimits(tt.args.load); (err != nil) != tt.wantErr {
				t.Errorf("withinLimits() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.load.Accepted != tt.wantAccepted {
				t.Errorf("withinLimits() accepted %v, want accepted %v", tt.args.load.Accepted, tt.wantAccepted)
			}
		})
	}
}
