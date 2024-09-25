use image::{GrayImage, ImageFormat};
use js_sys::{Promise, Uint8Array};
use std::io::Cursor;
use std::rc::Rc;
use wasm_bindgen::prelude::*;
use wasm_bindgen::JsCast;
use web_sys::{File, FileReader, ProgressEvent};

// 启用从 Rust 调用 JavaScript 的 console.log
macro_rules! console_log {
    ($($t:tt)*) => (web_sys::console::log_1(&format!($($t)*).into()))
}

// 异步读取文件并处理图像
#[wasm_bindgen]
pub fn process_image(file: File) -> Promise {
    console_log!("开始处理图像");

    let file = Rc::new(file);
    let promise = Promise::new(&mut |resolve, reject| {
        let reader = Rc::new(FileReader::new().unwrap());
        let file_clone = file.clone();
        let resolve_clone = resolve.clone();
        let reject_clone = reject.clone();

        let reader_clone = reader.clone();
        let onload = Closure::wrap(Box::new(move |_: ProgressEvent| {
            console_log!("FileReader 加载完成");
            let result = reader_clone.result().unwrap();
            let array = Uint8Array::new(&result);
            let mut buffer = vec![0; array.length() as usize];
            array.copy_to(&mut buffer[..]);

            // 处理图像
            match image::load_from_memory(&buffer) {
                Ok(img) => {
                    console_log!("图像解码成功");
                    // 转换为灰度图像
                    let gray_img: GrayImage = img.to_luma8();

                    // 编码为 PNG
                    let mut output_buffer = Vec::new();
                    let mut cursor = Cursor::new(&mut output_buffer);
                    match gray_img.write_to(&mut cursor, ImageFormat::Png) {
                        Ok(_) => {
                            console_log!("图像编码为 PNG 成功");
                            let uint8_array = Uint8Array::from(&output_buffer[..]);
                            resolve_clone.call1(&JsValue::NULL, &uint8_array).unwrap();
                        }
                        Err(e) => {
                            console_log!("图像编码失败");
                            reject_clone
                                .call1(
                                    &JsValue::NULL,
                                    &JsValue::from_str(&format!("编码失败: {}", e)),
                                )
                                .unwrap();
                        }
                    }
                }
                Err(e) => {
                    console_log!("图像解码失败");
                    reject_clone
                        .call1(
                            &JsValue::NULL,
                            &JsValue::from_str(&format!("解码失败: {}", e)),
                        )
                        .unwrap();
                }
            }
        }) as Box<dyn FnMut(_)>);

        reader.set_onload(Some(onload.as_ref().unchecked_ref()));
        onload.forget();

        // 读取文件为 ArrayBuffer
        match reader.read_as_array_buffer(&file_clone) {
            Ok(_) => {}
            Err(e) => {
                console_log!("读取文件失败");
            }
        }
    });

    promise
}
