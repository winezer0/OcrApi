#!/usr/bin/env python
# encoding: utf-8

import unittest
from unittest.mock import Mock, patch

import llm_ocr_api as service


class LlmOcrApiUnitTest(unittest.TestCase):
    """针对核心函数进行单元测试。"""

    def test_normalize_base64(self):
        """验证data-url和纯base64都可以被正确规范化。"""
        raw = "data:image/png;base64,QUJDRA=="
        self.assertEqual(service.normalize_base64(raw), "QUJDRA==")
        self.assertEqual(service.normalize_base64("  QUJDRA==  "), "QUJDRA==")

    def test_extract_result_number(self):
        """验证能从不同文本中提取最终数字。"""
        self.assertEqual(service.extract_result_number("结果是 15"), "15")
        self.assertEqual(service.extract_result_number("-9"), "-9")
        self.assertEqual(service.extract_result_number("6 + 3 = 9"), "9")
        self.assertEqual(service.extract_result_number("无数字"), "")

    def test_parse_request_base64_json(self):
        """验证JSON请求体中imageBase64字段可被识别。"""
        with service.APP.test_request_context(
            "/ruoyi/base64", method="POST", json={"imageBase64": "QUJDRA=="}
        ):
            self.assertEqual(service.parse_request_base64(service.request), "QUJDRA==")

    @patch("uitls.OpenAI")
    def test_call_dashscope_ocr(self, mock_openai):
        """验证百炼接口返回内容可以正确解析。"""
        mock_message = Mock()
        mock_message.content = "7"
        mock_choice = Mock()
        mock_choice.message = mock_message
        mock_completion = Mock()
        mock_completion.choices = [mock_choice]
        mock_client = Mock()
        mock_client.chat.completions.create.return_value = mock_completion
        mock_openai.return_value = mock_client

        output = service.call_dashscope_ocr("QUJDRA==", "sk-test")
        self.assertEqual(output, "7")
        self.assertTrue(mock_openai.called)


if __name__ == "__main__":
    unittest.main()
