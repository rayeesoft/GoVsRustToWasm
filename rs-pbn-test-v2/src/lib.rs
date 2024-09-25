use js_sys::Uint8Array;
use wasm_bindgen::prelude::*;

/// 将图像转换为灰度图像
///
/// # 参数
///
/// - `data`: 输入的图像数据，格式为 RGBA，每个像素 4 个字节
/// - `width`: 图像的宽度
/// - `height`: 图像的高度
///
/// # 返回
///
/// 返回处理后的图像数据，格式为 RGBA，每个像素 4 个字节
#[wasm_bindgen]
pub fn process_image(data: &[u8], width: u32, height: u32) -> Uint8Array {
    // 检查数据长度是否匹配
    if data.len() != (width * height * 4) as usize {
        // 数据长度不匹配，返回空的 Uint8Array
        return Uint8Array::new_with_length(0);
    }

    // 创建一个可变的副本来存储处理后的数据
    let mut processed_data = data.to_vec();

    // 遍历每个像素（每个像素 4 个字节：R、G、B、A）
    for i in (0..processed_data.len()).step_by(4) {
        let r = processed_data[i] as f32;
        let g = processed_data[i + 1] as f32;
        let b = processed_data[i + 2] as f32;

        // 使用标准亮度公式计算灰度值
        let gray = (0.299 * r + 0.587 * g + 0.114 * b) as u8;

        // 将 R、G、B 设置为灰度值，保持 A 不变
        processed_data[i] = gray;
        processed_data[i + 1] = gray;
        processed_data[i + 2] = gray;
        // Alpha (processed_data[i + 3]) 保持不变
    }

    // 将处理后的数据转换为 Uint8Array 并返回
    Uint8Array::from(&processed_data[..])
}
