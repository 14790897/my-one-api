import React, { useEffect, useState } from 'react';
import { API, showError } from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { Table, Tag } from '@douyinfe/semi-ui';
import { renderQuota } from '../helpers/render';

const TokensTable = () => {
  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
    },
    {
      title: '剩余额度',
      dataIndex: 'remain_quota',
      render: (text, record, index) => {
        return (
          <div>
            {record.unlimited_quota ? (
              <Tag size={'large'} color={'white'}>
                无限制
              </Tag>
            ) : (
              <Tag size={'large'} color={'light-blue'}>
                {renderQuota(parseInt(text))}
              </Tag>
            )}
          </div>
        );
      },
    },
  ];

  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [tokens, setTokens] = useState([]);
  const [tokenCount, setTokenCount] = useState(pageSize);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);

  const setTokensFormat = (tokens) => {
    setTokens(tokens);
    if (tokens.length >= pageSize) {
      setTokenCount(tokens.length + pageSize);
    } else {
      setTokenCount(tokens.length);
    }
  };

  let pageData = tokens.slice(
    (activePage - 1) * pageSize,
    activePage * pageSize,
  );
  const loadTokens = async (startIdx) => {
    setLoading(true);
    const res = await API.get(`/api/token/?p=${startIdx}&size=${pageSize}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setTokensFormat(data);
      } else {
        let newTokens = [...tokens];
        newTokens.splice(startIdx * pageSize, data.length, ...data);
        setTokensFormat(newTokens);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadTokens(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [pageSize]);

  const handlePageChange = (page) => {
    setActivePage(page);
    if (page === Math.ceil(tokens.length / pageSize) + 1) {
      // In this case we have to load more data and then append them.
      loadTokens(page - 1).then((r) => {});
    }
  };

  return (
    <Table
      style={{ marginTop: 20 }}
      columns={columns}
      dataSource={pageData}
      pagination={{
        currentPage: activePage,
        pageSize: pageSize,
        total: tokenCount,
        showSizeChanger: true,
        pageSizeOptions: [10, 20, 50, 100],
        formatPageText: (page) =>
          `第 ${page.currentStart} - ${page.currentEnd} 条，共 ${tokens.length} 条`,
        onPageSizeChange: (size) => {
          setPageSize(size);
          setActivePage(1);
        },
        onPageChange: handlePageChange,
      }}
      loading={loading}
    ></Table>
  );
};

export default TokensTable;
