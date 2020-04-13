// Libraries
import React, {PureComponent} from 'react'
import {withRouter, WithRouterProps} from 'react-router'
import memoizeOne from 'memoize-one'

// Components
import BucketCard from 'src/buckets/components/BucketCard'
import DemoDataBucketCard from 'src/buckets/components/DemoDataBucketCard'
import {ResourceList} from '@influxdata/clockface'

// Selectors
import {getSortedResources} from 'src/shared/utils/sort'

// Types
import {Bucket, OwnBucket} from 'src/types'
import {Sort} from '@influxdata/clockface'

// Utils
import {SortTypes} from 'src/shared/utils/sort'

type SortKey = keyof Bucket | 'retentionRules[0].everySeconds'

interface Props {
  buckets: Bucket[]
  emptyState: JSX.Element
  onUpdateBucket: (b: OwnBucket) => void
  onDeleteBucket: (b: OwnBucket) => void
  onFilterChange: (searchTerm: string) => void
  sortKey: string
  sortDirection: Sort
  sortType: SortTypes
  onClickColumn: (
    sortType: SortTypes
  ) => (nextSort: Sort, sortKey: SortKey) => void
}

class BucketList extends PureComponent<Props & WithRouterProps> {
  private memGetSortedResources = memoizeOne<typeof getSortedResources>(
    getSortedResources
  )

  public render() {
    const {sortKey, sortDirection, onClickColumn} = this.props
    return (
      <>
        <ResourceList>
          <ResourceList.Header>
            <ResourceList.Sorter
              name="Name"
              sortKey={this.headerKeys[0]}
              sort={sortKey === this.headerKeys[0] ? sortDirection : Sort.None}
              onClick={onClickColumn(SortTypes.String)}
              testID="name-sorter"
            />
            <ResourceList.Sorter
              name="Retention"
              sortKey={this.headerKeys[1]}
              sort={sortKey === this.headerKeys[1] ? sortDirection : Sort.None}
              onClick={onClickColumn(SortTypes.Float)}
              testID="retention-sorter"
            />
          </ResourceList.Header>
          <ResourceList.Body emptyState={this.props.emptyState}>
            {this.listBuckets}
          </ResourceList.Body>
        </ResourceList>
      </>
    )
  }

  private get headerKeys(): SortKey[] {
    return ['name', 'retentionRules[0].everySeconds']
  }

  private get listBuckets(): JSX.Element[] {
    const {
      buckets,
      sortKey,
      sortDirection,
      sortType,
      onDeleteBucket,
      onFilterChange,
      onUpdateBucket,
    } = this.props
    const sortedBuckets = this.memGetSortedResources(
      buckets,
      sortKey,
      sortDirection,
      sortType
    )

    return sortedBuckets.map(bucket => {
      if (bucket.type === 'demodata') {
        return <DemoDataBucketCard key={bucket.id} bucket={bucket} />
      }
      return (
        <BucketCard
          key={bucket.id}
          bucket={bucket}
          onDeleteBucket={onDeleteBucket}
          onDeleteData={this.handleStartDeleteData}
          onUpdateBucket={onUpdateBucket}
          onFilterChange={onFilterChange}
        />
      )
    })
  }

  private handleStartDeleteData = (bucket: OwnBucket) => {
    const {orgID} = this.props.params

    this.props.router.push(
      `/orgs/${orgID}/load-data/buckets/${bucket.id}/delete-data`
    )
  }
}

export default withRouter<Props>(BucketList)
