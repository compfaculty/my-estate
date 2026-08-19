package mock

func fixtures() []Payload {
	// Kyiv sale apartments. Two duplicate pairs (mock-001/001b, mock-002/002b).
	return []Payload{
		listing("mock-001", "Поділ", "Хорива", "12", 50.4691, 30.5134, 3, 75, 2, 5, 2680000,
			"Світла 3-кімнатна біля Контрактової. Нещодавно знижена ціна.", "Олена Коваль", "private", "0501111001", true, true, true, false, 2004, "good"),
		listing("mock-001b", "Поділ", "Хорива", "12", 50.4692, 30.5135, 3, 76, 2, 5, 2750000,
			"Той самий будинок, друге оголошення. 3 кімнати, 76 м².", "Поділ Homes", "broker", "0501111002", true, true, true, false, 2004, "good"),

		listing("mock-002", "Оболонь", "Оболонський проспект", "8", 50.5122, 30.4981, 4, 110, 6, 12, 4290000,
			"Високий поверх, 4 кімнати, парк і паркінг.", "Андрій Мельник", "private", "0502222001", true, true, true, true, 2012, "renovated"),
		listing("mock-002b", "Оболонь", "Оболонський проспект", "8", 50.5122, 30.4982, 4, 108, 6, 12, 4350000,
			"Брокерське оголошення оболонської 4-кімнатної.", "North Side Realty", "broker", "0502222002", true, true, true, true, 2012, "renovated"),

		listing("mock-003", "Печерськ", "Липська", "19", 50.4418, 30.5381, 2, 52, 1, 3, 3180000,
			"Затишна 2-кімнатна в старому будинку. Без ліфта.", "Ярина Мороз", "private", "0503333001", false, false, true, false, 1935, "renovated"),
		listing("mock-004", "Шевченківський", "Володимирська", "45", 50.4509, 30.5144, 3.5, 92, 4, 8, 5120000,
			"Класична 3.5 кімнати на Володимирській. Ліфт.", "White City Brokers", "broker", "0504444001", true, true, true, true, 1998, "good"),
		listing("mock-005", "Поділ", "Сагайдачного", "140", 50.4634, 30.5212, 3, 81, 3, 6, 3890000,
			"До Дніпра пішки. 3 кімнати, балкон, кондиціонер.", "Ніна Авраменко", "private", "0505555001", true, false, true, false, 1978, "needs renovation"),
		listing("mock-006", "Позняки", "Драгоманова", "88", 50.3981, 30.6336, 4, 120, 2, 4, 3550000,
			"Простора квартира на Позняках з високими стелями.", "Позняки Living", "broker", "0506666001", false, true, true, true, 1965, "good"),
		listing("mock-007", "Оболонь", "Героїв Дніпра", "22", 50.5224, 30.4982, 5, 145, 8, 15, 6200000,
			"Сімейна 5-кімнатна біля метро. Два паркомісця.", "Софія Ткач", "private", "0507777001", true, true, true, true, 2016, "renovated"),
		listing("mock-008", "Солом'янка", "Преображенська", "7", 50.4347, 30.4891, 3, 78, 3, 8, 2190000,
			"Доступна 3-кімнатна біля залізничного вокзалу.", "Солома Deals", "broker", "0508888001", true, true, false, true, 2001, "good"),
		listing("mock-009", "Голосіїв", "Васильківська", "56", 50.3934, 30.5022, 2.5, 61, 1, 4, 1980000,
			"Тиха 2.5 кімнати. Добрий попит на оренду.", "Майя Бар", "private", "0509999001", false, false, true, false, 1988, "needs renovation"),
		listing("mock-010", "Куренівка", "Кирилівська", "14", 50.4789, 30.4688, 3, 70, 2, 3, 3450000,
			"Біля ринку. Третій поверх без ліфта.", "Куренівка Properties", "broker", "0501010001", false, false, true, false, 1952, "good"),
		listing("mock-011", "Поділ", "Нижній Вал", "5", 50.4681, 30.5115, 4, 95, 3, 6, 2990000,
			"Відкрите планування, 4 кімнати, лофт-стиль.", "Олег Гадад", "private", "0501011001", true, false, true, true, 2008, "renovated"),
		listing("mock-012", "Оболонь", "Маршала Тимошенка", "33", 50.5081, 30.4948, 3, 88, 5, 10, 3650000,
			"Тиха вулиця, ліфт, комора.", "Оболонь Estates", "broker", "0501012001", true, true, true, true, 2006, "good"),
		listing("mock-013", "Печерськ", "Інститутська", "9", 50.4474, 30.5319, 5, 160, 2, 2, 8900000,
			"Рідкісний дуплекс біля Арсеналу.", "Приватний власник", "private", "0501013001", false, true, true, false, 1920, "renovated"),
		listing("mock-014", "Шевченківський", "Хрещатик", "21", 50.4478, 30.5221, 2, 48, 5, 7, 2450000,
			"Компактна квартира над магазинами.", "Center Brokers", "broker", "0501014001", true, false, false, false, 1960, "needs renovation"),
		listing("mock-015", "Поділ", "Велика Житомирська", "20", 50.4549, 30.5085, 3.5, 86, 4, 9, 4100000,
			"Балкон і паркінг у центрі.", "Тамара Азулай", "private", "0501015001", true, true, true, true, 1995, "good"),
		listing("mock-016", "Дарниця", "Харківське шосе", "12", 50.4398, 30.6289, 3, 82, 1, 5, 2720000,
			"Перший поверх з палісадником.", "Дарниця Homes", "broker", "0501016001", false, true, true, false, 1972, "good"),
		listing("mock-017", "Оболонь", "Північна", "15", 50.5282, 30.4984, 4, 118, 7, 12, 5150000,
			"Світла 4-кімнатна до зеленої зони.", "Гіл Бен", "private", "0501017001", true, true, true, true, 2010, "renovated"),
		listing("mock-018", "Солом'янка", "Урицького", "40", 50.4315, 30.4882, 2, 55, 4, 8, 1750000,
			"Стартова 2-кімнатна біля вокзалу.", "Masger Realty", "broker", "0501018001", true, false, false, true, 1992, "needs renovation"),
		listing("mock-019", "Голосіїв", "Голосіївський проспект", "88", 50.3972, 30.5088, 3, 74, 2, 5, 2280000,
			"Кутова квартира на дві сторони.", "Ігор Шараб", "private", "0501019001", false, true, true, true, 1985, "good"),
		listing("mock-020", "Куренівка", "Фрунзе", "3", 50.4811, 30.4679, 1.5, 38, 1, 2, 1890000,
			"Маленька студія над ринком.", "Market Flats", "broker", "0501020001", false, false, false, false, 1948, "needs renovation"),
		listing("mock-021", "Поділ", "Межигірська", "21", 50.4717, 30.5152, 3, 80, 6, 8, 2850000,
			"Високий поверх, краєвид, ліфт.", "Ліор Кац", "private", "0501021001", true, true, true, true, 2018, "renovated"),
		listing("mock-022", "Оболонь", "Прирічна", "4", 50.5182, 30.4991, 5, 132, 3, 7, 5480000,
			"Велика сімейна, два балкони.", "Park Obolon Ltd", "broker", "0501022001", true, true, true, true, 2009, "good"),
		listing("mock-023", "Печерськ", "Шовковична", "2", 50.4443, 30.5301, 2, 58, 3, 4, 3350000,
			"Дизайнерська 2-кімнатна, повний ремонт.", "Studio Pechersk", "broker", "0501023001", false, false, true, false, 1938, "renovated"),
		listing("mock-024", "Шевченківський", "Богдана Хмельницького", "18", 50.4488, 30.5072, 4, 105, 5, 11, 6400000,
			"Преміум 4-кімнатна біля оперного.", "Хіла Розен", "private", "0501024001", true, true, true, true, 2014, "renovated"),
		listing("mock-025", "Поділ", "Набережно-Хрещатицька", "48", 50.4635, 30.5259, 3, 79, 2, 5, 3720000,
			"Набережна, тихий двір.", "North Keys", "broker", "0501025001", true, false, true, false, 1975, "good"),
		listing("mock-026", "Позняки", "Ахматової", "1", 50.3947, 30.6388, 3.5, 98, 1, 3, 4100000,
			"Біля озера, цегляний будинок.", "Pozniaky RE", "broker", "0501026001", false, true, true, false, 1910, "renovated"),
		listing("mock-027", "Оболонь", "Лятошинського", "70", 50.5161, 30.4967, 4, 112, 9, 16, 4900000,
			"Вежа з паркінгом і коморою.", "Obolon Towers", "broker", "0501027001", true, true, true, true, 2003, "good"),
		listing("mock-028", "Солом'янка", "Чоколівський бульвар", "11", 50.4279, 30.4633, 3, 72, 1, 4, 2050000,
			"Перший поверх, смуга зелені.", "Kadesh Living", "broker", "0501028001", false, true, true, true, 1980, "good"),
		listing("mock-029", "Голосіїв", "Максимовича", "17", 50.3858, 30.4934, 4, 100, 4, 6, 2650000,
			"Перепланована 4-кімнатна.", "Роніт Алон", "private", "0501029001", true, true, true, true, 1999, "good"),
		listing("mock-030", "Куренівка", "Олени Теліги", "8", 50.4897, 30.4599, 2, 50, 2, 3, 2100000,
			"Буде знято з публікації на наступному циклі.", "Withdrawn Seller", "private", "0501030001", false, false, true, false, 1955, "needs renovation"),
	}
}

func listing(id, neighborhood, street, house string, lat, lng float64, rooms, area float64, floor, total int, price float64, desc, seller, sellerType, phone string, elevator, parking, balcony, safe bool, year int, reno string) Payload {
	bedrooms := int(rooms)
	if rooms != float64(bedrooms) {
		bedrooms = int(rooms)
	}
	if bedrooms > 0 {
		bedrooms--
	}
	photos := []string{
		"https://picsum.photos/seed/" + id + "a/800/600",
		"https://picsum.photos/seed/" + id + "b/800/600",
		"https://picsum.photos/seed/" + id + "c/800/600",
	}
	return Payload{
		SourceListingID:  id,
		SourceURL:        "https://mock.local/listings/" + id,
		City:             "Київ",
		District:         "Київ",
		Neighborhood:     neighborhood,
		Street:           street,
		HouseNumber:      house,
		Lat:              lat,
		Lng:              lng,
		PropertyType:     "apartment",
		Rooms:            rooms,
		Bedrooms:         bedrooms,
		Bathrooms:        1,
		Floor:            floor,
		TotalFloors:      total,
		AreaM2:           area,
		BalconyAreaM2:    8,
		ParkingSpaces:    boolInt(parking),
		ConstructionYear: year,
		RenovationState:  reno,
		Elevator:         elevator,
		Parking:          parking,
		Storage:          parking,
		Balcony:          balcony,
		AirConditioning:  true,
		SafeRoom:         safe,
		Accessibility:    elevator,
		ListingType:      "sale",
		Price:            price,
		Currency:         "UAH",
		Description:      desc,
		SellerName:       seller,
		SellerType:       sellerType,
		Phone:            phone,
		Photos:           photos,
		Status:           "active",
	}
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
