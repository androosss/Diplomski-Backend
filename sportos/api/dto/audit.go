package dto

import (
	"backend/sportos/repo/crud"
	DR "backend/sportos/repo/dto"
	"context"
	"time"
)

// [swagger]

// Audit
//
// Audit:
// swagger:model Audit
type Audit struct {
	//Audit unique id
	Id string
	//Id of database table
	Entity DR.SportosEntity `json:"entity,omitempty"`
	//Name of database table
	EntityName string `json:"entityName,omitempty"`
	//Id of entity that is changed
	EntityId string `json:"entityId,omitempty"`
	//What crud action was done
	CrudAction DR.CrudAction `json:"crudAction,omitempty"`
	//Old value of changed fields
	Old interface{} `json:"old,omitempty"`
	//New value of changed fields
	New interface{} `json:"new,omitempty"`
	//Source ip form where request was send
	SourceIp string `json:"sourceIp,omitempty"`
	//Date when audit was created
	CreatedAt time.Time `json:"createdAt,omitempty"`
	//User who created audit
	CreatedBy string `json:"createdBy,omitempty"`
}

func (p *Audit) InitSourceIp(ctx context.Context, Repo *crud.Repo, Id string) {
	if Id == "" {
		p.SourceIp = ""
	} else {
		apiJournal, _ := Repo.ApiJournalCrud.GetById(ctx, Id, nil)
		p.SourceIp = *apiJournal.SourceIP
	}
}

func (p *Audit) InitWithDatabaseStruct(do *DR.Audit) {
	p.Id = do.AuditId
	p.Entity = do.Entity
	p.EntityName = do.Entity.GetName()
	p.EntityId = do.EntityId
	p.CrudAction = *do.CrudAction
	p.Old = do.Old
	p.New = do.New
	p.CreatedAt = do.CreatedAt
	p.CreatedBy = do.CreatedBy
}
