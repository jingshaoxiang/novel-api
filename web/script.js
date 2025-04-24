// **请根据你的后端实际部署地址修改**
// 如果前端和后端在同一个域，并且后端在 /web 路径下提供前端文件，
// 理论上 API_BASE_URL 可以设置为 "", "/tokens/upload", "/tokens/count" 等相对路径。
// 但为了清晰，仍然推荐使用完整的 URL 或相对路径前缀如 "/api"
const API_BASE_URL = "/"; // 假设 API 接口直接挂在根路径下，例如 /tokens/upload
// 如果你的 API 在 /api 前缀下，可以设置为 "/api"

// 获取当前可用 Tokens 数量
async function getAvailableTokensCount() {
    const countElement = document.getElementById('available-tokens-count');
    try {
        // 假设后端有一个 GET /tokens/count 接口返回可用数量
        const response = await fetch(`${API_BASE_URL}tokens/count`); // 使用相对路径
        if (response.ok) {
            const data = await response.json();
            countElement.textContent = data.count; // 假设后端返回 { count: 4 }
        } else {
            console.error("Failed to get available tokens count:", response.statusText);
            countElement.textContent = "获取失败";
        }
    } catch (error) {
        console.error("Error fetching tokens count:", error);
        countElement.textContent = "连接错误";
    }
}

// 上传密钥列表到服务器
async function uploadTokens() {
    const tokensTextarea = document.getElementById('tokens');
    const tokensContent = tokensTextarea.value;
    const tokensArray = tokensContent.split('\n').map(line => line.trim()).filter(line => line !== ""); // 按行分割，去除首尾空格和空行

    try {
        // 假设后端有一个 POST /tokens/upload 接口，接收一个包含 tokens 数组的 JSON
        const response = await fetch(`${API_BASE_URL}tokens/upload`, { // 使用相对路径
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                // 如果后端需要认证，可以在这里添加认证头
            },
            body: JSON.stringify({ tokens: tokensArray })
        });

        if (response.ok) {
            alert("密钥已成功上传！");
            // 上传成功后更新可用 Tokens 数量
            getAvailableTokensCount();
            // 清空错误 Tokens 显示
            document.getElementById('error-tokens-display').style.display = 'none';
        } else {
            const errorData = await response.text(); // 尝试获取后端返回的错误信息
            alert(`上传失败: ${response.statusText}${errorData ? ' - ' + errorData : ''}`);
            console.error("Upload failed:", response.status, errorData);
        }
    } catch (error) {
        console.error("Error uploading tokens:", error);
        alert("上传过程中发生错误，请检查网络或后端服务。");
    }
}

// 查看错误 Tokens（需要后端提供相关接口）
async function viewErrorTokens() {
    const errorDisplay = document.getElementById('error-tokens-display');
    errorDisplay.style.display = 'none'; // 先隐藏之前的错误信息
    errorDisplay.textContent = ''; // 清空之前的内容

    try {
        // 假设后端有一个 GET /tokens/errors 接口返回错误密钥列表
        const response = await fetch(`${API_BASE_URL}tokens/errors`); // 使用相对路径
        if (response.ok) {
            const data = await response.json(); // 假设返回 { errors: ["invalid_token1", "invalid_token2"] }
            if (data.errors && Array.isArray(data.errors) && data.errors.length > 0) {
                errorDisplay.textContent = "错误 Tokens:\n" + data.errors.join('\n');
                errorDisplay.style.display = 'block';
            } else {
                errorDisplay.textContent = "没有错误 Tokens。";
                errorDisplay.style.display = 'block';
            }
        } else {
            console.error("Failed to get error tokens:", response.statusText);
            errorDisplay.textContent = `获取错误 Tokens 失败: ${response.statusText}`;
            errorDisplay.style.display = 'block';
        }
    } catch (error) {
        console.error("Error fetching error tokens:", error);
        errorDisplay.textContent = "获取错误 Tokens 过程中发生错误。";
        errorDisplay.style.display = 'block';
    }
}

// 清空所有 Tokens
async function clearTokens() {
    if (confirm("确定要清空所有 Tokens 吗？此操作不可撤销！")) {
        try {
            // 假设后端有一个 DELETE /tokens 接口用于清空
            const response = await fetch(`${API_BASE_URL}tokens`, { // 使用相对路径
                method: 'DELETE',
                // 如果后端需要认证，可以在这里添加认证头
            });

            if (response.ok) {
                alert("所有 Tokens 已成功清空！");
                // 清空成功后更新可用 Tokens 数量
                getAvailableTokensCount();
                // 清空文本框内容
                document.getElementById('tokens').value = '';
                // 隐藏错误 Tokens 显示
                document.getElementById('error-tokens-display').style.display = 'none';
            } else {
                const errorData = await response.text();
                alert(`清空失败: ${response.statusText}${errorData ? ' - ' + errorData : ''}`);
                console.error("Clear failed:", response.status, errorData);
            }
        } catch (error) {
            console.error("Error clearing tokens:", error);
            alert("清空过程中发生错误，请检查网络或后端服务。");
        }
    }
}

// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', (event) => {
    getAvailableTokensCount();
    // 你可以在这里添加代码来加载当前的 Tokens 列表到文本框，如果后端提供了相应的接口
    // loadTokensFromServer(); // 假设你实现了这个函数
});