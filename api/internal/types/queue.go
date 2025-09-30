package types

import "time"

type Entry struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	WaitingRoomID string    `bson:"waitingRoomId" json:"waitingRoomId"`
	TicketNumber  string    `bson:"ticketNumber" json:"ticketNumber"`
	QRToken       string    `bson:"qrToken" json:"qrToken"`
	Status        string    `bson:"status" json:"status"` // WAITING, CALLED, IN_SERVICE, COMPLETED, SKIPPED, CANCELLED, NO_SHOW
	Position      int       `bson:"position" json:"position"`
	CreatedAt     time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
	CardData      CardData  `bson:"cardData,omitempty" json:"cardData,omitempty"`
}

type CardData struct {
	IDNumber    string `bson:"idNumber" json:"idNumber"`
	FirstName   string `bson:"firstName" json:"firstName"`
	LastName    string `bson:"lastName" json:"lastName"`
	DateOfBirth string `bson:"dateOfBirth" json:"dateOfBirth"`
	Gender      string `bson:"gender" json:"gender"`
	Nationality string `bson:"nationality" json:"nationality"`
	Address     string `bson:"address" json:"address"`
	IssuedDate  string `bson:"issuedDate" json:"issuedDate"`
	ExpiryDate  string `bson:"expiryDate" json:"expiryDate"`
	Photo       string `bson:"photo" json:"photo"`
	Source      string `bson:"source" json:"source"`
}
