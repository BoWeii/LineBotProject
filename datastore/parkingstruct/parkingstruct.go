package parkingstruct

//ParkingLotNTPC 新北市停車場
type ParkingLotNTPC struct {
	ID          string  `json:"id"`                    //停車場序號
	Name        string  `json:"name"`                  //停車場名稱
	Type        int     `json:"type,string"`           //1：剩餘車位數 2：靜態停車場資料
	Tel         string  `json:"tel"`                   //停車場電話
	Address     string  `json:"address" datastore:"-"` //地址，不儲存
	Pay         string  `json:"payEx"`                 //停車場收費資訊
	ServiceTime string  `json:"serviceTime"`           //服務時間
	TotalCar    int     `json:"totalCar,string"`       //總汽車數
	TotalMotor  int     `json:"totalMotor,string"`     //總機車數
	Lat         float64 `json:"twd97Y,string"`         //緯度
	Lon         float64 `json:"twd97X,string"`         //經度
}
