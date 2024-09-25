package main

import (
	"math"
)

// 常量定义
const (
	RED              = iota // 红色通道
	GREEN                   // 绿色通道
	BLUE                    // 蓝色通道
	MAX_COLOR        = 256
	SIGNIFICANT_BITS = 5
	MAX_SIDE_INDEX   = 1 << SIGNIFICANT_BITS
	SIDE_SIZE        = MAX_SIDE_INDEX + 1
)

// getIndex 根据红绿蓝值计算索引
func getIndex(r, g, b int) int {
	return r*SIDE_SIZE*SIDE_SIZE + g*SIDE_SIZE + b
}

// ColorCube 代表一个颜色立方体
type ColorCube struct {
	RedMin, RedMax     int
	GreenMin, GreenMax int
	BlueMin, BlueMax   int
	Volume             int
}

// NewColorCube 创建一个新的 ColorCube 实例
func NewColorCube() *ColorCube {
	return &ColorCube{
		RedMin:   0,
		RedMax:   0,
		GreenMin: 0,
		GreenMax: 0,
		BlueMin:  0,
		BlueMax:  0,
		Volume:   0,
	}
}

// Uint8Array2D 储存每个像素对应的颜色索引
type Uint8Array2D struct {
	Width  uint32
	Height uint32
	Data   []uint8
}

// NewUint8Array2D 创建一个新的 Uint8Array2D 实例
func NewUint8Array2D(width, height uint32) *Uint8Array2D {
	return &Uint8Array2D{
		Width:  width,
		Height: height,
		Data:   make([]uint8, width*height),
	}
}

// Set 设置指定位置的值
func (ua *Uint8Array2D) Set(x, y uint32, value uint8) {
	if x < ua.Width && y < ua.Height {
		index := y*ua.Width + x
		ua.Data[index] = value
	}
}

// Get 获取指定位置的值
func (ua *Uint8Array2D) Get(x, y uint32) uint8 {
	if x < ua.Width && y < ua.Height {
		index := y*ua.Width + x
		return ua.Data[index]
	}
	return 0
}

// SetByIndex 根据线性索引设置值
func (ua *Uint8Array2D) SetByIndex(index int, value uint8) {
	if index >= 0 && index < len(ua.Data) {
		ua.Data[index] = value
	}
}

// Clone 克隆 Uint8Array2D
func (ua *Uint8Array2D) Clone() *Uint8Array2D {
	dataCopy := make([]uint8, len(ua.Data))
	copy(dataCopy, ua.Data)
	return &Uint8Array2D{
		Width:  ua.Width,
		Height: ua.Height,
		Data:   dataCopy,
	}
}

// Bitmap 代表一张位图
type Bitmap struct {
	Width  uint32
	Height uint32
	Data   []uint8 // RGBA格式
}

// NewBitmap 创建一个新的 Bitmap 实例
func NewBitmap(width, height uint32) *Bitmap {
	return &Bitmap{
		Width:  width,
		Height: height,
		Data:   make([]uint8, width*height*4), // 每个像素4个字节
	}
}

// NewBitmap 创建一个新的 Bitmap 实例
func NewBitmapWithData(width, height uint32, data []uint8) *Bitmap {
	return &Bitmap{
		Width:  width,
		Height: height,
		Data:   data, // 每个像素4个字节
	}
}

// Clone 克隆一个 Bitmap
func (bmp *Bitmap) Clone() *Bitmap {
	dataCopy := make([]uint8, len(bmp.Data))
	copy(dataCopy, bmp.Data)
	return &Bitmap{
		Width:  bmp.Width,
		Height: bmp.Height,
		Data:   dataCopy,
	}
}

// ColorMap 储存颜色映射信息
type ColorMap struct {
	Width         uint32
	Height        uint32
	Colors        [][4]uint8    // 储存所有颜色，索引即为颜色索引
	MappedIndices *Uint8Array2D // 储存每个像素对应的颜色索引
}

// NewColorMap 创建一个新的 ColorMap 实例
func NewColorMap(width, height uint32, colors [][4]uint8) *ColorMap {
	return &ColorMap{
		Width:         width,
		Height:        height,
		Colors:        colors,
		MappedIndices: NewUint8Array2D(width, height),
	}
}

// SetPixelIndex 设置指定像素的颜色索引
func (cm *ColorMap) SetPixelIndex(x, y uint32, index uint8) {
	cm.MappedIndices.Set(x, y, index)
}

// GetPixelIndex 获取指定像素的颜色索引
func (cm *ColorMap) GetPixelIndex(x, y uint32) (uint8, bool) {
	if x < cm.Width && y < cm.Height {
		return cm.MappedIndices.Get(x, y), true
	}
	return 0, false
}

// GetPixelColor 获取指定像素的颜色
func (cm *ColorMap) GetPixelColor(x, y uint32) ([4]uint8, bool) {
	index, ok := cm.GetPixelIndex(x, y)
	if !ok {
		return [4]uint8{}, false
	}
	if int(index) < len(cm.Colors) {
		return cm.Colors[index], true
	}
	return [4]uint8{}, false
}

// ToImage 将 ColorMap 转换为 Bitmap
func (cm *ColorMap) ToImage() *Bitmap {
	bitmap := NewBitmap(cm.Width, cm.Height)

	if len(cm.MappedIndices.Data) != int(cm.Width*cm.Height) {
		panic("mapped_indices 的长度与 Bitmap 的像素数不匹配")
	}

	for i, index := range cm.MappedIndices.Data {
		var color [4]uint8
		if int(index) < len(cm.Colors) {
			color = cm.Colors[index]
		} else {
			color = [4]uint8{0, 0, 0, 255} // 默认颜色
		}

		dataPos := i * 4
		if dataPos+3 < len(bitmap.Data) {
			bitmap.Data[dataPos] = color[0]
			bitmap.Data[dataPos+1] = color[1]
			bitmap.Data[dataPos+2] = color[2]
			bitmap.Data[dataPos+3] = color[3]
		}
	}

	return bitmap
}

// Quantizer 颜色量化器
type Quantizer struct {
	Colors       int
	Weights      []float64
	MomentsRed   []float64
	MomentsGreen []float64
	MomentsBlue  []float64
	Moments      []float64
	Table        [256]float64
	Cubes        []*ColorCube
	Palette      [][4]uint8
	Bitmap       *Bitmap
}

// NewQuantizer 创建一个新的 Quantizer 实例
func NewQuantizer(bitmap *Bitmap, colors int) *Quantizer {
	if colors > MAX_COLOR {
		colors = MAX_COLOR
	}

	totalSize := SIDE_SIZE * SIDE_SIZE * SIDE_SIZE

	weights := make([]float64, totalSize)
	momentsRed := make([]float64, totalSize)
	momentsGreen := make([]float64, totalSize)
	momentsBlue := make([]float64, totalSize)
	moments := make([]float64, totalSize)

	// 预计算平方表
	var table [256]float64
	for i := 0; i < 256; i++ {
		table[i] = float64(i * i)
	}

	// 初始化颜色立方体
	cubes := make([]*ColorCube, colors+1)
	for i := 0; i < len(cubes); i++ {
		cubes[i] = NewColorCube()
	}

	// 初始化调色板
	palette := make([][4]uint8, colors)
	for i := 0; i < colors; i++ {
		palette[i] = [4]uint8{0, 0, 0, 0}
	}

	// 正确初始化cubes[0]
	cubes[0].RedMin = 0
	cubes[0].RedMax = SIDE_SIZE - 1
	cubes[0].GreenMin = 0
	cubes[0].GreenMax = SIDE_SIZE - 1
	cubes[0].BlueMin = 0
	cubes[0].BlueMax = SIDE_SIZE - 1
	cubes[0].Volume = (cubes[0].RedMax - cubes[0].RedMin) *
		(cubes[0].GreenMax - cubes[0].GreenMin) *
		(cubes[0].BlueMax - cubes[0].BlueMin)

	quant := &Quantizer{
		Colors:       colors,
		Weights:      weights,
		MomentsRed:   momentsRed,
		MomentsGreen: momentsGreen,
		MomentsBlue:  momentsBlue,
		Moments:      moments,
		Table:        table,
		Cubes:        cubes,
		Palette:      palette,
		Bitmap:       bitmap.Clone(),
	}

	quant.sample(bitmap.Data)
	return quant
}

// sample 采样像素数据，pixels是RGBA格式的字节数组
func (q *Quantizer) sample(pixels []uint8) {
	for i := 0; i < len(pixels); i += 4 {
		if i+2 >= len(pixels) {
			break
		}
		r, g, b := pixels[i], pixels[i+1], pixels[i+2]
		q.addColor([3]uint8{r, g, b})
	}
}

// addColor 将单个RGB颜色添加到权重和矩中
func (q *Quantizer) addColor(color [3]uint8) {
	bitsToRemove := 8 - SIGNIFICANT_BITS
	indexRed := int(color[0]>>bitsToRemove) + 1
	indexGreen := int(color[1]>>bitsToRemove) + 1
	indexBlue := int(color[2]>>bitsToRemove) + 1

	index := getIndex(indexRed, indexGreen, indexBlue)

	q.Weights[index] += 1.0
	q.MomentsRed[index] += float64(color[0])
	q.MomentsGreen[index] += float64(color[1])
	q.MomentsBlue[index] += float64(color[2])
	q.Moments[index] += q.Table[color[0]] + q.Table[color[1]] + q.Table[color[2]]
}

// Quantize 执行量化并返回量化后的 Bitmap
func (q *Quantizer) Quantize() *Bitmap {
	q.BuildPalette()
	return q.MapPixels().ToImage()
}

// BuildPalette 执行量化过程，返回调色板（RGBA格式，Alpha固定为255）
func (q *Quantizer) BuildPalette() [][4]uint8 {
	q.calculateMoments()
	palette := q.preparePalette()
	q.Palette = palette
	return palette
}

// calculateMoments 计算积分矩
func (q *Quantizer) calculateMoments() {
	for r := 1; r < SIDE_SIZE; r++ {
		for g := 1; g < SIDE_SIZE; g++ {
			for b := 1; b < SIDE_SIZE; b++ {
				index := getIndex(r, g, b)
				indexR1 := getIndex(r-1, g, b)
				indexG1 := getIndex(r, g-1, b)
				indexB1 := getIndex(r, g, b-1)
				indexRG1 := getIndex(r-1, g-1, b)
				indexRB1 := getIndex(r-1, g, b-1)
				indexGB1 := getIndex(r, g-1, b-1)
				indexRGB1 := getIndex(r-1, g-1, b-1)

				q.Weights[index] += q.Weights[indexR1] + q.Weights[indexG1] + q.Weights[indexB1] -
					q.Weights[indexRG1] - q.Weights[indexRB1] - q.Weights[indexGB1] + q.Weights[indexRGB1]

				q.MomentsRed[index] += q.MomentsRed[indexR1] + q.MomentsRed[indexG1] + q.MomentsRed[indexB1] -
					q.MomentsRed[indexRG1] - q.MomentsRed[indexRB1] - q.MomentsRed[indexGB1] + q.MomentsRed[indexRGB1]

				q.MomentsGreen[index] += q.MomentsGreen[indexR1] + q.MomentsGreen[indexG1] + q.MomentsGreen[indexB1] -
					q.MomentsGreen[indexRG1] - q.MomentsGreen[indexRB1] - q.MomentsGreen[indexGB1] + q.MomentsGreen[indexRGB1]

				q.MomentsBlue[index] += q.MomentsBlue[indexR1] + q.MomentsBlue[indexG1] + q.MomentsBlue[indexB1] -
					q.MomentsBlue[indexRG1] - q.MomentsBlue[indexRB1] - q.MomentsBlue[indexGB1] + q.MomentsBlue[indexRGB1]

				q.Moments[index] += q.Moments[indexR1] + q.Moments[indexG1] + q.Moments[indexB1] -
					q.Moments[indexRG1] - q.Moments[indexRB1] - q.Moments[indexGB1] + q.Moments[indexRGB1]
			}
		}
	}
}

// preparePalette 准备调色板，返回RGBA格式的颜色数组，Alpha固定为255
func (q *Quantizer) preparePalette() [][4]uint8 {
	next := 0
	volumeVariance := make([]float64, q.Colors+1) // +1 to prevent index out of range

	for i := 1; i < q.Colors; i++ {
		if next >= len(q.Cubes) {
			break
		}
		cubeNext := q.Cubes[next]
		cubeI := q.Cubes[i]

		if !q.cut(cubeNext, cubeI) {
			volumeVariance[next] = 0.0
			continue
		}

		if cubeNext.Volume > 1 {
			volumeVariance[next] = q.calculateVariance(cubeNext)
		} else {
			volumeVariance[next] = 0.0
		}

		if cubeI.Volume > 1 {
			volumeVariance[i] = q.calculateVariance(cubeI)
		} else {
			volumeVariance[i] = 0.0
		}

		// 选择具有最大方差的立方体进行下一步分割
		next = 0
		temp := volumeVariance[0]
		for k := 1; k <= i; k++ {
			if k >= len(volumeVariance) {
				break
			}
			if volumeVariance[k] > temp {
				temp = volumeVariance[k]
				next = k
			}
		}

		if temp <= 0.0 {
			q.Colors = i + 1
			break
		}
	}

	// 生成调色板
	palette := make([][4]uint8, 0, q.Colors)

	for k := 0; k < q.Colors; k++ {
		weight := q.volume(q.Cubes[k], q.Weights)
		if weight > 0.0 {
			r := q.volume(q.Cubes[k], q.MomentsRed) / weight
			g := q.volume(q.Cubes[k], q.MomentsGreen) / weight
			b := q.volume(q.Cubes[k], q.MomentsBlue) / weight

			// Clamp values between 0 and 255
			r = math.Min(math.Max(r, 0.0), 255.0)
			g = math.Min(math.Max(g, 0.0), 255.0)
			b = math.Min(math.Max(b, 0.0), 255.0)

			palette = append(palette, [4]uint8{
				uint8(r),
				uint8(g),
				uint8(b),
				255, // Alpha固定为255
			})
		}
	}

	return palette
}

// cut 分割颜色立方体，更新 first 和 second
func (q *Quantizer) cut(first, second *ColorCube) bool {
	wholeRed := q.volume(first, q.MomentsRed)
	wholeGreen := q.volume(first, q.MomentsGreen)
	wholeBlue := q.volume(first, q.MomentsBlue)
	wholeWeight := q.volume(first, q.Weights)

	// 在每个颜色通道上寻找最佳切割位置
	maxRed, cutRed := q.maximize(first, RED, wholeRed, wholeGreen, wholeBlue, wholeWeight)
	maxGreen, cutGreen := q.maximize(first, GREEN, wholeRed, wholeGreen, wholeBlue, wholeWeight)
	maxBlue, cutBlue := q.maximize(first, BLUE, wholeRed, wholeGreen, wholeBlue, wholeWeight)

	// 确定哪个颜色通道的切割效果最好
	direction := RED
	max := maxRed
	cut := cutRed

	if maxGreen > max {
		max = maxGreen
		direction = GREEN
		cut = cutGreen
	}
	if maxBlue > max {
		// max = maxBlue
		direction = BLUE
		cut = cutBlue
	}

	if cut < 0 {
		return false
	}

	// 根据切割方向更新两个立方体
	*second = *first
	second.RedMax = first.RedMax
	second.GreenMax = first.GreenMax
	second.BlueMax = first.BlueMax

	switch direction {
	case RED:
		first.RedMax = cut
		second.RedMin = first.RedMax
		second.GreenMin = first.GreenMin
		second.BlueMin = first.BlueMin
	case GREEN:
		first.GreenMax = cut
		second.GreenMin = first.GreenMax
		second.RedMin = first.RedMin
		second.BlueMin = first.BlueMin
	case BLUE:
		first.BlueMax = cut
		second.BlueMin = first.BlueMax
		second.RedMin = first.RedMin
		second.GreenMin = first.GreenMin
	}

	// 更新体积
	first.Volume = (first.RedMax - first.RedMin) *
		(first.GreenMax - first.GreenMin) *
		(first.BlueMax - first.BlueMin)
	second.Volume = (second.RedMax - second.RedMin) *
		(second.GreenMax - second.GreenMin) *
		(second.BlueMax - second.BlueMin)

	return true
}

// maximize 在指定方向上寻找最佳切割位置
func (q *Quantizer) maximize(cube *ColorCube, direction int, wholeRed, wholeGreen, wholeBlue, wholeWeight float64) (float64, int) {
	max := 0.0
	cutPosition := -1

	minPos := q.cubeMin(cube, direction) + 1
	maxPos := q.cubeMax(cube, direction)

	for position := minPos; position < maxPos; position++ {
		halfRed := q.bottom(cube, direction, q.MomentsRed) + q.top(cube, direction, position, q.MomentsRed)
		halfGreen := q.bottom(cube, direction, q.MomentsGreen) + q.top(cube, direction, position, q.MomentsGreen)
		halfBlue := q.bottom(cube, direction, q.MomentsBlue) + q.top(cube, direction, position, q.MomentsBlue)
		halfWeight := q.bottom(cube, direction, q.Weights) + q.top(cube, direction, position, q.Weights)

		if halfWeight == 0.0 {
			continue
		}

		halfDistance := (halfRed*halfRed + halfGreen*halfGreen + halfBlue*halfBlue) / halfWeight

		remainingRed := wholeRed - halfRed
		remainingGreen := wholeGreen - halfGreen
		remainingBlue := wholeBlue - halfBlue
		remainingWeight := wholeWeight - halfWeight

		if remainingWeight == 0.0 {
			continue
		}

		remainingDistance := (remainingRed*remainingRed + remainingGreen*remainingGreen + remainingBlue*remainingBlue) / remainingWeight

		temp := halfDistance + remainingDistance

		if temp > max {
			max = temp
			cutPosition = position
		}
	}

	return max, cutPosition
}

// volume 计算指定立方体在某个矩上的体积
func (q *Quantizer) volume(cube *ColorCube, moment []float64) float64 {
	res := moment[getIndex(cube.RedMax, cube.GreenMax, cube.BlueMax)] -
		moment[getIndex(cube.RedMax, cube.GreenMax, cube.BlueMin)] -
		moment[getIndex(cube.RedMax, cube.GreenMin, cube.BlueMax)] +
		moment[getIndex(cube.RedMax, cube.GreenMin, cube.BlueMin)] -
		moment[getIndex(cube.RedMin, cube.GreenMax, cube.BlueMax)] +
		moment[getIndex(cube.RedMin, cube.GreenMax, cube.BlueMin)] +
		moment[getIndex(cube.RedMin, cube.GreenMin, cube.BlueMax)] -
		moment[getIndex(cube.RedMin, cube.GreenMin, cube.BlueMin)]
	return res
}

// top 计算指定立方体在切割方向上的上半部分
func (q *Quantizer) top(cube *ColorCube, direction, position int, moment []float64) float64 {
	switch direction {
	case RED:
		return moment[getIndex(position, cube.GreenMax, cube.BlueMax)] -
			moment[getIndex(position, cube.GreenMax, cube.BlueMin)] -
			moment[getIndex(position, cube.GreenMin, cube.BlueMax)] +
			moment[getIndex(position, cube.GreenMin, cube.BlueMin)]
	case GREEN:
		return moment[getIndex(cube.RedMax, position, cube.BlueMax)] -
			moment[getIndex(cube.RedMax, position, cube.BlueMin)] -
			moment[getIndex(cube.RedMin, position, cube.BlueMax)] +
			moment[getIndex(cube.RedMin, position, cube.BlueMin)]
	case BLUE:
		return moment[getIndex(cube.RedMax, cube.GreenMax, position)] -
			moment[getIndex(cube.RedMax, cube.GreenMin, position)] -
			moment[getIndex(cube.RedMin, cube.GreenMax, position)] +
			moment[getIndex(cube.RedMin, cube.GreenMin, position)]
	default:
		return 0.0
	}
}

// bottom 计算指定立方体在切割方向上的下半部分
func (q *Quantizer) bottom(cube *ColorCube, direction int, moment []float64) float64 {
	switch direction {
	case RED:
		return -moment[getIndex(cube.RedMin, cube.GreenMax, cube.BlueMax)] +
			moment[getIndex(cube.RedMin, cube.GreenMax, cube.BlueMin)] +
			moment[getIndex(cube.RedMin, cube.GreenMin, cube.BlueMax)] -
			moment[getIndex(cube.RedMin, cube.GreenMin, cube.BlueMin)]
	case GREEN:
		return -moment[getIndex(cube.RedMax, cube.GreenMin, cube.BlueMax)] +
			moment[getIndex(cube.RedMax, cube.GreenMin, cube.BlueMin)] +
			moment[getIndex(cube.RedMin, cube.GreenMin, cube.BlueMax)] -
			moment[getIndex(cube.RedMin, cube.GreenMin, cube.BlueMin)]
	case BLUE:
		return -moment[getIndex(cube.RedMax, cube.GreenMax, cube.BlueMin)] +
			moment[getIndex(cube.RedMax, cube.GreenMin, cube.BlueMin)] +
			moment[getIndex(cube.RedMin, cube.GreenMax, cube.BlueMin)] -
			moment[getIndex(cube.RedMin, cube.GreenMin, cube.BlueMin)]
	default:
		return 0.0
	}
}

// calculateVariance 计算立方体内的颜色方差
func (q *Quantizer) calculateVariance(cube *ColorCube) float64 {
	volumeRed := q.volume(cube, q.MomentsRed)
	volumeGreen := q.volume(cube, q.MomentsGreen)
	volumeBlue := q.volume(cube, q.MomentsBlue)
	volumeMoment := q.volume(cube, q.Moments)
	weight := q.volume(cube, q.Weights)

	distance := volumeRed*volumeRed + volumeGreen*volumeGreen + volumeBlue*volumeBlue

	return volumeMoment - (distance / weight)
}

// cubeMin 获取立方体在指定方向上的最小值
func (q *Quantizer) cubeMin(cube *ColorCube, direction int) int {
	switch direction {
	case RED:
		return cube.RedMin
	case GREEN:
		return cube.GreenMin
	case BLUE:
		return cube.BlueMin
	default:
		return 0
	}
}

// cubeMax 获取立方体在指定方向上的最大值
func (q *Quantizer) cubeMax(cube *ColorCube, direction int) int {
	switch direction {
	case RED:
		return cube.RedMax
	case GREEN:
		return cube.GreenMax
	case BLUE:
		return cube.BlueMax
	default:
		return 0
	}
}

// MapPixels 映射原始图像像素到量化后的图像
func (q *Quantizer) MapPixels() *ColorMap {
	width := q.Bitmap.Width
	height := q.Bitmap.Height
	colors := q.Palette
	mappedIndices := NewUint8Array2D(width, height)

	index := 0
	for i := 0; i < len(q.Bitmap.Data); i += 4 {
		if i+3 >= len(q.Bitmap.Data) {
			break
		}
		r, g, b, a := q.Bitmap.Data[i], q.Bitmap.Data[i+1], q.Bitmap.Data[i+2], q.Bitmap.Data[i+3]
		colorIndex := q.findClosestPaletteIndex([4]uint8{r, g, b, a})
		mappedIndices.SetByIndex(index, colorIndex)
		index++
	}

	return &ColorMap{
		Width:         width,
		Height:        height,
		Colors:        colors,
		MappedIndices: mappedIndices,
	}
}

// findClosestPaletteIndex 找到最接近的调色板颜色的索引
func (q *Quantizer) findClosestPaletteIndex(color [4]uint8) uint8 {
	minDistance := uint32(math.MaxUint32)
	closestIndex := uint8(0)

	for idx, paletteColor := range q.Palette {
		distance := q.colorDistanceSquared(color, paletteColor)
		if distance < minDistance {
			minDistance = distance
			closestIndex = uint8(idx)
		}
	}

	return closestIndex
}

// colorDistanceSquared 计算两个颜色之间的欧几里得距离的平方
func (q *Quantizer) colorDistanceSquared(color1, color2 [4]uint8) uint32 {
	rDiff := int32(color1[0]) - int32(color2[0])
	gDiff := int32(color1[1]) - int32(color2[1])
	bDiff := int32(color1[2]) - int32(color2[2])
	aDiff := int32(color1[3]) - int32(color2[3])

	return uint32(rDiff*rDiff + gDiff*gDiff + bDiff*bDiff + aDiff*aDiff)
}
