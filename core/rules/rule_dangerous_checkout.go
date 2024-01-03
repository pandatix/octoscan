package rules

import (
	"octoscan/common"
	"strings"

	"github.com/rhysd/actionlint"
)

type RuleDangerousCheckout struct {
	actionlint.RuleBase
	checkoutPos *actionlint.Pos
}

// NewRuleDangerousCheckout creates new RuleDangerousCheckout instance.
func NewRuleDangerousCheckout() *RuleDangerousCheckout {
	return &RuleDangerousCheckout{
		RuleBase: actionlint.NewRuleBase(
			"dangerous-checkout",
			"Check for dangerous checkout.",
		),
		checkoutPos: nil,
	}
}

// VisitStep is callback when visiting Step node.
func (rule *RuleDangerousCheckout) VisitStep(n *actionlint.Step) error {
	// For now we only trigger the rule once
	// might be a good idea to trigger the rule for each checkout
	if rule.checkoutPos != nil {
		return nil
	}

	e, ok := n.Exec.(*actionlint.ExecAction)
	if !ok || e.Uses == nil {
		return nil
	}

	if e.Uses.ContainsExpression() {
		// Cannot parse specification made with interpolation. Give up
		return nil
	}

	spec := e.Uses.Value

	// search for checkout action
	if strings.HasPrefix(spec, "actions/checkout") {
		// basicRegExp := regexp.MustCompile(`github.event.pull_request`)
		ref := e.Inputs["ref"]

		if ref != nil {
			rule.checkoutPos = e.Uses.Pos
		}
	}

	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children
func (rule *RuleDangerousCheckout) VisitWorkflowPost(n *actionlint.Workflow) error {
	if rule.checkoutPos != nil {

		for _, event := range n.On {
			if common.IsStringInArray(common.TriggerWithExternalData, event.EventName()) {
				rule.Errorf(
					rule.checkoutPos,
					"Use of checkout action with %q workflow trigger and custom ref.",
					event.EventName(),
				)

				// only trigger once even if both trigger are defined
				return nil
			}
		}
	}

	return nil
}
