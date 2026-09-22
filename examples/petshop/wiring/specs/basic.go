// Package specs defines PetShop deployment wiring.
package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/petshop/workflow/petshop"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
)

var Basic = cmdbuilder.SpecOption{
	Name: "basic", Description: "A single-process PetShop deployment using Blueprint in-memory backends", Build: makeBasic,
}

func makeBasic(spec wiring.WiringSpec) ([]string, error) {
	petDB := simple.NoSQLDB(spec, "pet_db")
	petSearch := workflow.Service[petshop.PetSearchService](spec, "pet_search_service", petDB, "")

	historyQueue := simple.Queue(spec, "adoption_history_queue")
	payForAdoption := workflow.Service[petshop.PayForAdoptionService](spec, "pay_for_adoption_service", petDB, petSearch, historyQueue)
	petListAdoptions := workflow.Service[petshop.PetListAdoptionsService](spec, "pet_list_adoptions_service", petDB, petSearch)

	foodDB := simple.NoSQLDB(spec, "pet_food_db")
	foodCart := simple.Cache(spec, "pet_food_cart")
	petFood := workflow.Service[petshop.PetFoodService](spec, "pet_food_service", foodDB, foodCart)
	petFoodAgent := workflow.Service[petshop.PetFoodAgentService](spec, "pet_food_agent_service", petSearch, petFood)
	petSite := workflow.Service[petshop.PetSiteService](spec, "pet_site_service", petSearch, payForAdoption, petListAdoptions, petFood, petFoodAgent)

	return []string{petSite, petSearch, payForAdoption, petListAdoptions, petFood, petFoodAgent}, nil
}
