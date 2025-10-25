package goo

import (
	"chain"
	"chain/runtime"
)

func resolve(id string) int64 {
	dispute := getDispute(id)
	request := getRequest(id)
	if dispute.NbResolvedVotes == 0 {
		// If no one voted or reveal their vote, the proposed value wins by default
		return request.ProposedValue
	}
	if request.YesNoQuestion {
		return resolveYesNo(dispute)
	} else {
		return resolveNumeric(dispute)
	}
}

func resolveYesNo(dispute Dispute) int64 {
	yesVotes := int64(0)
	noVotes := int64(0)
	for _, vote := range dispute.Votes {
		if vote.Revealed {
			if vote.Value == 1 {
				yesVotes += vote.TokenAmount
			} else if vote.Value == 0 {
				noVotes += vote.TokenAmount
			}
		}
	}
	var winningValue int64
	var weight int64
	if yesVotes > noVotes {
		winningValue = 1
		weight = yesVotes
	} else {
		winningValue = 0
		weight = noVotes
	}
	rewardAndSlachVoters(dispute, winningValue, weight)
	return winningValue
}

// resolveNumeric determines the winning value for numeric disputes based on the highest total token weight.
// Not the best algorithm, but gas effective and fast.
func resolveNumeric(dispute Dispute) int64 {
	// Collect unique values and aggregate weights
	valueWeights := make(map[int64]int64)
	for _, vote := range dispute.Votes {
		if vote.Revealed {
			valueWeights[vote.Value] += vote.TokenAmount
		}
	}

	// Find the value with the most token weight
	var winningValue int64
	var maxWeight int64

	for value, weight := range valueWeights {
		if weight > maxWeight {
			maxWeight = weight
			winningValue = value
		}
	}
	rewardAndSlachVoters(dispute, winningValue, maxWeight)
	return winningValue
}

func rewardAndSlachVoters(dispute Dispute, winningValue, totalWeight int64) {
	for _, vote := range dispute.Votes {
		if vote.Revealed {
			if vote.Value == winningValue {
				// Reward winning voters
				VoteToken.mint(vote.Voter, 2)
				reward := bond * (vote.TokenAmount / totalWeight)
				if reward != 0 {
					bank.SendCoins(runtime.CurrentRealm().Address(), vote.Voter, chain.Coins{chain.Coin{Denom: "ugnot", Amount: reward}})
				}
				chain.Emit("VoterRewarded", "voter", vote.Voter.String())
			} else {
				// Slash losing voters
				slash := vote.TokenAmount / 4 // 25% slash
				VoteToken.burn(vote.Voter, slash)
				chain.Emit("VoterSlashed", "voter", vote.Voter.String())
			}
		}
	}
}
