package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/govcd"
)

var TestAccVcdCatalogItem = "TestAccVcdCatalogItemBasic"
var TestAccVcdCatalogItemDescription = "TestAccVcdCatalogItemBasicDescription"

func TestAccVcdCatalogItemBasic(t *testing.T) {

	var catalogItem govcd.CatalogItem
	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Catalog":         testConfig.VCD.Catalog.Name,
		"CatalogItemName": TestAccVcdCatalogItem,
		"Description":     TestAccVcdCatalogItemDescription,
		"OvaPath":         testConfig.Ova.OvaPath,
		"UploadPieceSize": testConfig.Ova.UploadPieceSize,
		"UploadProgress":  testConfig.Ova.UploadProgress,
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	configText := templateFill(testAccCheckVcdCatalogItemBasic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCatalogItemDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestAccVcdCatalogItem, &catalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "name", TestAccVcdCatalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "description", TestAccVcdCatalogItemDescription),
				),
			},
		},
	})
}

func testAccCheckVcdCatalogItemExists(itemName string, catalogItem *govcd.CatalogItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		catalogItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if catalogItemRs.Primary.ID == "" {
			return fmt.Errorf("no catalog item ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := adminOrg.FindCatalog(testConfig.VCD.Catalog.Name)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%#v)", testConfig.VCD.Catalog.Name, catalog.Catalog)
		}

		newCatalogItem, err := catalog.FindCatalogItem(catalogItemRs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("catalog item %s does not exist (%#v)", catalogItemRs.Primary.ID, catalogItem.CatalogItem)
		}

		catalogItem = &newCatalogItem
		return nil
	}
}

func testAccCheckCatalogItemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_catalog_item" && rs.Primary.Attributes["name"] != TestAccVcdCatalogItem {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := adminOrg.FindCatalog(testConfig.VCD.Catalog.Name)
		if err != nil {
			return fmt.Errorf("catalog query %s ended with error: %#v", rs.Primary.ID, err)
		}

		itemName := rs.Primary.Attributes["name"]
		catalogItem, err := catalog.FindCatalogItem(itemName)

		if catalogItem != (govcd.CatalogItem{}) {
			return fmt.Errorf("catalog item %s still exists", itemName)
		}
		if err != nil {
			return fmt.Errorf("catalog item %s still exists or other error: %#v", itemName, err)
		}

	}

	return nil
}

const testAccCheckVcdCatalogItemBasic = `
  resource "vcd_catalog_item" "{{.CatalogItemName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogItemName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"
}
`
