package goo

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strconv"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
)

type DataRequest struct {
	Id             string
	Creator        address
	Timestamp      time.Time
	AncillaryData  string
	YesNoQuestion  bool
	ProposedValue  int64
	Proposer       address
	ProposerBond   int64
	Disputer       address
	DisputerBond   int64
	ResolutionTime time.Time
	WinningValue   int64
	State          string // "Requested", "Proposed", "Disputed", "Resolved"
	Deadline       time.Time
}

var (
	admin           = address("g1qdwemgqgqm7s42gy6xcj7f22uhfdzr82st3sy3")
	ResolutionTime  = 2 * int64(time.Minute.Seconds())
	RequesterReward = 1 * int64(1_000_000) // in GNOT
	Bond            = 2 * int64(1_000_000) // in GNOT
	Requests        = avl.NewTree()
	Bank            = banker.NewBanker(banker.BankerTypeRealmSend)
	idGenerator     seqid.ID
)

// -- PUBLIC FUNCTIONS --

// RequestData allows a user to request data from the oracle.
// `RequesterReward` value needs to be sent to the contract as a reward
// You need to ask a question that can be answered with a single number like a yes/no question (0 or 1) or a specific value (e.g. ETH/USD price).
func RequestData(cur realm, ancillaryData string, yesNoQuestion bool, deadline int64) string {
	if ancillaryData == "" {
		panic("error: Ancillary data cannot be empty.")
	}
	if deadline < time.Now().Add(24*time.Hour).Unix() {
		panic("error: Deadline must be at least 24 hours in the future.")
	}

	// TODO: check that x GNOT is sent as reward
	coins := banker.OriginSend()
	if len(coins) != 1 || coins.AmountOf("ugnot") != RequesterReward {
		panic("error: Incorrect reward amount sent. Required: " + strconv.FormatInt(RequesterReward, 10) + " ugnot.")
	}

	id := idGenerator.Next().String()

	request := DataRequest{
		Id:            id,
		Timestamp:     time.Now(),
		AncillaryData: ancillaryData,
		YesNoQuestion: yesNoQuestion,
		State:         "Requested",
		Deadline:      time.Unix(deadline, 0),
		Creator:       runtime.PreviousRealm().Address(),
	}
	Requests.Set(id, request)
	chain.Emit("DataRequested", "id", id, "timestamp", request.Timestamp.String(), "ancillaryData", ancillaryData)
	return id
}

// ProposeValue allows a user to propose a value for a requested data point.
// `Bond` value needs to be sent to the contract as a bond.
func ProposeValue(cur realm, id string, proposedValue int64) {
	request := getRequest(id)
	if request.State != "Requested" {
		panic("error: Request is not in 'Requested' state.")
	}
	if request.YesNoQuestion && proposedValue != 0 && proposedValue != 1 {
		panic("error: Proposed value must be 0 or 1 for yes/no questions.")
	}
	if time.Now().After(request.Deadline) {
		panic("error: Deadline for proposal has passed.")
	}

	coins := banker.OriginSend()
	if len(coins) != 1 || coins.AmountOf("ugnot") != Bond {
		panic("error: Incorrect bond amount sent. Required: " + strconv.FormatInt(Bond, 10) + " ugnot")
	}

	request.ProposedValue = proposedValue
	request.Proposer = runtime.PreviousRealm().Address()
	request.ProposerBond = Bond
	request.ResolutionTime = time.Now().Add(time.Duration(ResolutionTime) * time.Second)
	request.State = "Proposed"
	Requests.Set(id, request)
	chain.Emit("ValueProposed", "id", id, "proposedValue", strconv.Itoa(int(proposedValue)), "proposer", request.Proposer.String(), "resolutionTime", request.ResolutionTime.String())
}

// DisputeData allows a user to dispute a proposed value.
// `Bond` value needs to be sent to the contract as a bond.
func DisputeData(cur realm, id string) string {
	request := getRequest(id)
	if request.Proposer == runtime.PreviousRealm().Address() {
		panic("error: Proposer cannot dispute their own proposal.")
	}
	if request.State != "Proposed" {
		panic("error: Request is not in 'Proposed' state.")
	}
	if time.Now().After(request.ResolutionTime) {
		panic("error: Dispute period has ended.")
	}

	coins := banker.OriginSend()
	if len(coins) != 1 || coins.AmountOf("ugnot") != Bond {
		panic("error: Incorrect bond amount sent. Required: " + strconv.FormatInt(Bond, 10) + " ugnot")
	}

	request.Disputer = runtime.PreviousRealm().Address()
	request.DisputerBond = Bond
	request.State = "Disputed"
	Requests.Set(id, request)
	initiateDispute(id)
	chain.Emit("DataDisputed", "id", id, "disputer", request.Disputer.String())
	return "Time: " + time.Now().String()
}

// ResolveRequest finalizes an undisputed request after the resolution period has passed.
func ResolveRequest(cur realm, id string) {
	request := getRequest(id)
	if request.State == "Disputed" {
		panic("error: Request is in 'Disputed' state.")
	}
	if request.State == "Proposed" && request.ResolutionTime.After(time.Now()) {
		panic("error: Resolution period has not ended yet.")
	}
	if request.State == "Requested" {
		panic("error: Request has not been proposed yet.")
	}
	if request.State == "Resolved" {
		panic("error: Request is already resolved.")
	}
	request.State = "Resolved"
	request.WinningValue = request.ProposedValue
	Requests.Set(id, request)

	from := runtime.CurrentRealm().Address()
	to := request.Proposer
	totalPayout := request.ProposerBond + RequesterReward
	payout := chain.Coins{chain.Coin{Denom: "ugnot", Amount: totalPayout}}
	Bank.SendCoins(from, to, payout)

	chain.Emit("RequestResolved", "id", id, "winningValue", strconv.Itoa(int(request.WinningValue)))
}

// RequestResult returns the winning value of a resolved request.
func RequestResult(cur realm, id string) int64 {
	request := getRequest(id)
	if request.State != "Resolved" {
		panic("error: Request is not resolved.")
	}
	return request.WinningValue
}

// RequesterRetreiveFund allows the original requester to get their reward back if the deadline passed without a proposal.
func RequesterRetreiveFund(cur realm, id string) {
	request := getRequest(id)
	if request.State != "Requested" {
		panic("error: cannot retreive fund as requests fulfilled.")
	}
	if request.Creator != runtime.PreviousRealm().Address() {
		panic("error: Only the creator of the request can retreive the fund.")
	}
	if request.Deadline.After(time.Now()) {
		panic("error: Cannot retreive fund before the deadline.")
	}

	from := runtime.CurrentRealm().Address()
	to := request.Creator
	refund := chain.Coins{chain.Coin{Denom: "ugnot", Amount: RequesterReward}}
	Bank.SendCoins(from, to, refund)

	request.State = "Expired"
	Requests.Set(id, request)
	chain.Emit("RequestExpired", "id", id)
}

// -- ADMIN FUNCTIONS --

// SetResolutionDuration sets the duration (in seconds) for the resolution period.
func SetResolutionDuration(_ realm, duration int64) {
	if runtime.OriginCaller() == admin {
		ResolutionTime = duration
		chain.Emit("ResolutionTimeSet", "duration", strconv.Itoa(int(duration)))
	} else {
		panic("error: Only the admin can set the resolution time.")
	}
}

// SetRequesterReward sets the reward amount for a successful proposal.
func SetRequesterReward(_ realm, reward int64) {
	if runtime.OriginCaller() == admin {
		RequesterReward = reward
		chain.Emit("RequesterRewardSet", "reward", strconv.Itoa(int(reward)))
	} else {
		panic("error: Only the admin can set the requester reward.")
	}
}

// SetBond sets the bond amount required for proposals and disputes.
func SetBond(_ realm, bond int64) {
	if runtime.OriginCaller() == admin {
		Bond = bond
		chain.Emit("BondSet", "bond", strconv.Itoa(int(bond)))
	} else {
		panic("error: Only the admin can set the proposer bond.")
	}
}

// ChangeAdmin transfers admin privileges to a new address.
func ChangeAdmin(_ realm, newAdmin address) {
	if runtime.OriginCaller() == admin {
		admin = newAdmin
		chain.Emit("AdminChanged", "newAdmin", newAdmin.String())
	} else {
		panic("error: Only the admin can change the admin.")
	}
}

// -- VIEW FUNCTIONS --

// GetRequest returns the details of a specific data request.
func GetRequest(_ realm, id string) DataRequest {
	return *getRequest(id)
}

// GetBond returns the current bond amount.
func GetBond(_ realm) int64 {
	return Bond
}

// GetResolutionTime returns the current resolution time duration.
func GetResolutionTime(_ realm) int64 {
	return ResolutionTime
}

// GetRequesterReward returns the current requester reward amount.
func GetRequesterReward(_ realm) int64 {
	return RequesterReward
}

// GetRequestState returns the current state of a specific data request.
func GetRequestState(_ realm, id string) string {
	request := getRequest(id)
	return request.State
}

// GetAdmin returns the current admin address.
func GetAdmin(_ realm) address {
	return admin
}


// Utils functions

func getRequest(id string) *DataRequest {
	request, exists := Requests.Get(id)
	if !exists {
		panic("error: Request with this ID does not exist.")
	}
	return request.(*DataRequest)
}