package parkingstruct

//ParkingSpaceNTPC 新北市單一車格
type ParkingSpaceNTPC struct {
	ID            int     `json:"ID,string"`            //車格序號
	CELLID        float64 `json:"CELLID,string"`        //收費時段判斷
	Name          string  `json:"NAME"`                 //車格類型
	Day           string  `json:"DAY"`                  //收費天
	Hour          string  `json:"Hour"`                 //收費時段
	Pay           string  `json:"PAY"`                  //收費形式
	PayCash       string  `json:"PAYCASH"`              //費率
	Memo          string  `json:"MEMO"`                 //車格備註
	RoadID        string  `json:"ROADID"`               //路段代碼
	CellStatus    bool    `json:"CELLSTATUS,string"`    //車格狀態判斷 Y有車 N空位
	IsNowCash     bool    `json:"ISNOWCASH,string"`     //收費時段判斷
	ParkingStatus int     `json:"ParkingStatus,string"` //車格狀態 　1：有車、2：空位、3：非收費時段、4：時段性禁停、5：施工（民眾申請施工租用車格時使用）
	Lat           float64 `json:"lat,string"`           //緯度
	Lon           float64 `json:"lon,string"`           //經度
}

//ParkingLotNTPC 新北市停車場
type ParkingLotNTPC struct {
	ID          int     `json:"id,string"`             //停車場序號
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

//ParkingLotAvailNTPC 新北市停車場剩餘數量
type ParkingLotAvailNTPC struct {
	ID           int `json:"id,string"`           //停車場序號
	AvailableCar int `json:"availableCar,string"` //剩餘數量
}

//NTPC 新北市車格
type NTPC struct {
	Spaces []*ParkingSpaceNTPC
	Lot    []*ParkingLotNTPC
}

//FeeInfo 繳費資訊
type FeeInfo struct {
	TicketNo     string //收費編號
	CarID        string //車牌號碼
	Parkdt       string //開單日
	Paylim       string //繳費截止日
	AmountTicket string //費用
	CarType      string //車種
}
