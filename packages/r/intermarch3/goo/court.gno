package goo

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"gno.land/p/nt/avl"
)

var (
	Disputes        = avl.NewTree()
	DisputeDuration = 2 * int64(time.Minute.Seconds())
	RevealDuration  = 2 * int64(time.Minute.Seconds())
	VoteTokenPrice  = 1 * int64(1_000_000) // in GNOT
	VoteToken       = newOOToken("Gno Optimistic Oracle Token", "goot", 6)
)

func initiateDispute(id string) {
	if _, exists := Disputes.Get(id); exists {
		panic("error: Dispute for this request already exists.")
	}
	dispute := Dispute{
		RequestId:     id,
		Votes:         []Vote{},
		Voters:        avl.NewTree(),
		IsResolved:    false,
		EndTime:       time.Now().Add(time.Duration(DisputeDuration) * time.Second),
		EndRevealTime: time.Now().Add(time.Duration(DisputeDuration+RevealDuration) * time.Second),
	}
	Disputes.Set(id, dispute)
	chain.Emit("DisputeInitiated", "id", id)
}

// -- PUBLIC FUNCTIONS --

// BuyInitialVoteToken allows a user to buy their first vote token by sending VoteTokenPrice amount of ugnot.
func BuyInitialVoteToken(_ realm) {
	caller := runtime.OriginCaller()
	coins := banker.OriginSend()
	if len(coins) != 1 || coins.AmountOf("ugnot") != VoteTokenPrice {
		panic("error: Must send exactly " + strconv.Itoa(int(VoteTokenPrice/1_000_000)) + " gnot to get a vote token.")
	}

	balance := VoteToken.BalanceOf(caller)
	if balance > 0 {
		panic("error: You already have a vote token.")
	}

	VoteToken.Mint(caller, 1)
	chain.Emit("VoteTokenPurchased", "voter", caller.String())
}

// BalanceOfVoteToken returns the number of vote tokens held by the caller.
func BalanceOfVoteToken(_ realm) int64 {
	return VoteToken.balanceOf(runtime.PreviousRealm().Address())
}

// VoteOnDispute allows a user to commit a vote during a dispute.
func VoteOnDispute(cur realm, id string, hash string) {
	dispute := getDispute(id)
	res, _ := Requests.Get(id)
	request := res.(DataRequest)
	if request.Proposer == runtime.PreviousRealm().Address() || request.Disputer == runtime.PreviousRealm().Address() {
		panic("error: Proposer and Disputer cannot vote in this dispute.")
	}
	if dispute.IsResolved {
		panic("error: Dispute is already resolved.")
	}
	if time.Now().After(dispute.EndTime) {
		panic("error: Vote period has ended.")
	}
	amount := VoteToken.BalanceOf(runtime.PreviousRealm().Address())
	if amount < 1 {
		panic("error: You need at least 1 vote token to vote.")
	}

	vote := Vote{
		RequestId:   id,
		Voter:       runtime.PreviousRealm().Address(),
		TokenAmount: amount,
		Hash:        hash,
		Revealed:    false,
	}
	voter, exist := dispute.Voters.Get(string(vote.Voter))
	if exist && voter.(*Voter).HasVoted {
		panic("error: Voter has already voted in this dispute.")
	}
	dispute.Votes = append(dispute.Votes, vote)
	dispute.Voters.Set(string(vote.Voter), Voter{HasVoted: true, VoteIndex: int64(len(dispute.Votes) - 1)})
	Disputes.Set(id, dispute)
	chain.Emit("VoteSubmitted", "id", id, "voter", vote.Voter.String())
}

// RevealVote allows a user to reveal their vote after the voting period has ended.
func RevealVote(cur realm, id string, value int64, salt string) {
	dispute := getDispute(id)
	if dispute.IsResolved {
		panic("error: Dispute is resolved.")
	}
	if time.Now().Before(dispute.EndTime) {
		panic("error: Vote period has not ended yet.")
	}
	if time.Now().After(dispute.EndRevealTime) {
		panic("error: Reveal period has ended.")
	}
	voter, exist := dispute.Voters.Get(string(runtime.PreviousRealm().Address()))
	if !exist || !voter.(Voter).HasVoted {
		panic("error: Voter did not participate in this dispute.")
	}
	vote := dispute.Votes[voter.(Voter).VoteIndex]
	if vote.Revealed {
		panic("error: Vote already revealed.")
	}

	// Verify the hash
	res := sha256.Sum256([]byte(strconv.FormatInt(value, 10) + salt))
	expectedHash := hex.EncodeToString(res[:])
	if vote.Hash != expectedHash {
		panic("error: Hash does not match the revealed value and salt.")
	}
	vote.Value = value
	vote.Revealed = true
	dispute.NbResolvedVotes += 1
	dispute.Votes[voter.(Voter).VoteIndex] = vote
	Disputes.Set(id, dispute)
	chain.Emit("VoteRevealed", "id", id, "voter", vote.Voter.String(), "value", strconv.Itoa(int(value)))
}

// ResolveDispute finalizes a dispute after the reveal period, tallying votes and setting the winning value.
func ResolveDispute(cur realm, id string) {
	dispute := getDispute(id)
	if dispute.IsResolved {
		panic("error: Dispute is already resolved.")
	}
	if time.Now().Before(dispute.EndTime) {
		panic("error: Dispute period has not ended yet.")
	}
	val := resolve(id)
	dispute.WinningValue = val
	dispute.IsResolved = true
	Disputes.Set(id, dispute)
	// Update the original request with the winning value
	request := getRequest(id)

	request.ProposedValue = val
	request.State = "Resolved"
	Requests.Set(id, request)
	chain.Emit("DisputeResolved", "id", id, "winningValue", strconv.Itoa(int(val)))
	chain.Emit("RequestResolved", "id", id, "winningValue", strconv.Itoa(int(val)))

	var winner address
	if val != request.ProposedValue {
		// Refund + reward the disputer if the dispute changed the value
		winner = request.Disputer
	} else {
		// Refund + reward the proposer if the dispute did not change the value
		winner = request.Proposer
	}
	Bank.SendCoins(runtime.CurrentRealm().Address(), winner, chain.Coins{chain.Coin{Denom: "ugnot", Amount: Bond + RequesterReward}})
}

// -- admin functions --

// SetDisputeDuration sets the duration (in seconds) for the voting period.
func SetDisputeDuration(_ realm, duration int64) {
	if runtime.OriginCaller() == admin {
		DisputeDuration = duration * int64(time.Second)
	} else {
		panic("error: Only admin can set dispute duration.")
	}
}

// SetRevealDuration sets the duration (in seconds) for the reveal period.
func SetRevealDuration(_ realm, duration int64) {
	if runtime.OriginCaller() == admin {
		RevealDuration = duration * int64(time.Second)
	} else {
		panic("error: Only admin can set reveal duration.")
	}
}

// SetVoteTokenPrice sets the price (in ugnot) to cast a vote.
func SetVoteTokenPrice(_ realm, price int64) {
	if runtime.OriginCaller() == admin {
		VoteTokenPrice = price
	} else {
		panic("error: Only admin can set vote price.")
	}
}

// -- view functions --

// GetDispute returns the details of a specific dispute.
func GetDispute(_ realm, id string) Dispute {
	return getDispute(id)
}

// GetDisputeDuration returns the current dispute duration.
func GetDisputeDuration(_ realm) int64 {
	return DisputeDuration
}

// GetVoteTokenPrice returns the current vote price.
func GetVoteTokenPrice(_ realm) int64 {
	return VoteTokenPrice
}

// GetDisputeEndTime returns the end time of the voting period for a specific dispute.
func GetDisputeEndTime(_ realm, id string) time.Time {
	dispute := getDispute(id)
	return dispute.EndTime
}

// GetDisputeVotesAmount returns the total number of votes cast in a dispute.
func GetDisputeVotesAmount(_ realm, id string) int64 {
	dispute := getDispute(id)
	return int64(len(dispute.Votes))
}

// GetRevealEndTime returns the end time of the reveal period for a specific dispute.
func GetRevealEndTime(_ realm, id string) time.Time {
	dispute := getDispute(id)
	return dispute.EndRevealTime
}

// GetRevealDuration returns the current reveal duration.
func GetRevealDuration(_ realm) int64 {
	return RevealDuration
}

// Utils functions

func getDispute(id string) Dispute {
	dispute, exists := Disputes.Get(id)
	if !exists {
		panic("error: Dispute with this ID does not exist.")
	}
	return dispute.(Dispute)
}
